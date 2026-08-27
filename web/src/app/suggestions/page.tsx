"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Nav } from "@/components/nav";
import { apiFetch, getAccessToken } from "@/lib/api";

type SuggestionPayload = {
  target_exercise_id?: string;
  delta_percent?: number;
  suggested_exercise_name?: string;
  note?: string;
};

type AISuggestion = {
  id: string;
  client_name: string;
  kind: string;
  payload: SuggestionPayload;
  rationale: string;
  confidence?: number;
  status: string;
  created_at: string;
};

const KIND_LABELS: Record<string, string> = {
  volume_adjust: "Ajustar volumen",
  load_adjust: "Ajustar carga",
  exercise_swap: "Cambiar ejercicio",
  deload: "Semana de descarga",
  rest_day: "Dia de descanso",
  habit_nudge: "Empujon de habito",
};

function describePayload(p: SuggestionPayload): string {
  const parts: string[] = [];
  if (p.delta_percent !== undefined) {
    const sign = p.delta_percent > 0 ? "+" : "";
    parts.push(`${sign}${p.delta_percent}%`);
  }
  if (p.suggested_exercise_name) parts.push(`Sugerido: ${p.suggested_exercise_name}`);
  if (p.note) parts.push(p.note);
  return parts.join(" · ") || "Sin detalle adicional";
}

export default function SuggestionsPage() {
  const router = useRouter();
  const [suggestions, setSuggestions] = useState<AISuggestion[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [reviewingId, setReviewingId] = useState<string | null>(null);

  function load() {
    apiFetch("/v1/coach/suggestions")
      .then(async (res) => {
        if (!res.ok) throw new Error("No se pudo cargar las sugerencias");
        setSuggestions(await res.json());
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Error desconocido"));
  }

  useEffect(() => {
    if (!getAccessToken()) {
      router.replace("/login");
      return;
    }
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [router]);

  async function review(id: string, action: "approve" | "reject") {
    setReviewingId(id);
    try {
      const res = await apiFetch(`/v1/ai-suggestions/${id}/${action}`, { method: "POST" });
      if (!res.ok) throw new Error("No se pudo registrar la decision");
      setSuggestions((prev) => prev?.filter((s) => s.id !== id) ?? null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error desconocido");
    } finally {
      setReviewingId(null);
    }
  }

  return (
    <div className="min-h-screen bg-neutral-50 p-6 dark:bg-neutral-950 sm:p-10">
      <div className="mx-auto flex max-w-2xl flex-col gap-6">
        <Nav />
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Sugerencias de IA</h1>
          <p className="text-sm text-muted-foreground">
            Nunca se aplican solas — revisa cada una antes de aprobarla.
          </p>
        </div>

        {error && <p className="text-sm text-red-500">{error}</p>}

        {!suggestions && !error && (
          <p className="text-sm text-muted-foreground">Cargando...</p>
        )}

        {suggestions && suggestions.length === 0 && (
          <p className="text-sm text-muted-foreground">No hay sugerencias pendientes.</p>
        )}

        {suggestions?.map((s) => (
          <Card key={s.id}>
            <CardHeader className="flex flex-row items-start justify-between gap-2">
              <div>
                <CardTitle className="text-base">
                  {KIND_LABELS[s.kind] ?? s.kind} — {s.client_name}
                </CardTitle>
                <p className="text-sm text-muted-foreground">{describePayload(s.payload)}</p>
              </div>
              {s.confidence !== undefined && (
                <Badge variant="secondary">{Math.round(s.confidence * 100)}% seguro</Badge>
              )}
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              <p className="text-sm">{s.rationale}</p>
              <div className="flex gap-3">
                <Button
                  className="flex-1"
                  variant="outline"
                  disabled={reviewingId === s.id}
                  onClick={() => review(s.id, "reject")}
                >
                  Rechazar
                </Button>
                <Button
                  className="flex-1"
                  disabled={reviewingId === s.id}
                  onClick={() => review(s.id, "approve")}
                >
                  Aprobar
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
