"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Nav } from "@/components/nav";
import { apiFetch, getAccessToken } from "@/lib/api";

type CoachClient = {
  client_user_id: string;
  full_name: string;
  assignment_name?: string;
  has_assignment: boolean;
  last_session_at?: string;
  streak_days: number;
  recent_prs: number;
  needs_attention: boolean;
};

type PersonalRecord = {
  exercise_id: string;
  type: string;
  value: number;
  achieved_at: string;
};

type Exercise = { id: string; name: string };

const RECORD_LABEL: Record<string, string> = {
  max_weight: "Peso máximo",
  max_reps: "Reps máximas",
  estimated_1rm: "1RM estimado",
  max_volume_set: "Volumen máx. por serie",
};

function formatLastSession(iso?: string): string {
  if (!iso) return "Nunca entrenó";
  const days = Math.floor((Date.now() - new Date(iso).getTime()) / 86_400_000);
  if (days <= 0) return "Hoy";
  if (days === 1) return "Ayer";
  return `Hace ${days} días`;
}

export default function ClientDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const [client, setClient] = useState<CoachClient | null>(null);
  const [records, setRecords] = useState<PersonalRecord[] | null>(null);
  const [catalog, setCatalog] = useState<Exercise[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!getAccessToken()) {
      router.replace("/login");
      return;
    }
    apiFetch("/v1/coach/clients")
      .then(async (res) => {
        if (!res.ok) throw new Error("No se pudo cargar el cliente");
        const all: CoachClient[] = await res.json();
        setClient(all.find((c) => c.client_user_id === id) ?? null);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Error desconocido"));
    apiFetch(`/v1/coach/clients/${id}/progress`)
      .then(async (res) => {
        if (!res.ok) throw new Error("No se pudo cargar el progreso");
        setRecords(await res.json());
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Error desconocido"));
    apiFetch("/v1/exercises?limit=100")
      .then((res) => res.json())
      .then(setCatalog)
      .catch(() => setCatalog([]));
  }, [id, router]);

  const exerciseName = (exerciseId: string) =>
    catalog.find((e) => e.id === exerciseId)?.name ?? exerciseId.slice(0, 8);

  // Los records vienen planos; se agrupan por ejercicio para que el coach lea
  // "cómo va este movimiento", que es la pregunta real.
  const byExercise = new Map<string, PersonalRecord[]>();
  for (const r of records ?? []) {
    byExercise.set(r.exercise_id, [...(byExercise.get(r.exercise_id) ?? []), r]);
  }

  return (
    <div className="min-h-screen bg-neutral-50 p-6 dark:bg-neutral-950 sm:p-10">
      <div className="mx-auto flex max-w-4xl flex-col gap-6">
        <Nav />

        <div>
          <a href="/dashboard" className="text-sm text-muted-foreground underline underline-offset-4">
            ← Volver a clientes
          </a>
          <h1 className="mt-1 text-2xl font-bold tracking-tight">{client?.full_name ?? "Cliente"}</h1>
        </div>

        {error && <p className="text-sm text-red-500">{error}</p>}

        {client && (
          <div className="grid gap-3 sm:grid-cols-4">
            <StatCard label="Plan" value={client.has_assignment ? (client.assignment_name ?? "—") : "Sin plan"} />
            <StatCard label="Último entreno" value={formatLastSession(client.last_session_at)} />
            <StatCard label="Racha" value={`${client.streak_days} días`} />
            <StatCard label="PRs (14d)" value={String(client.recent_prs)} />
          </div>
        )}

        {client?.needs_attention && (
          <Badge variant="destructive" className="w-fit">
            Necesita atención — sin entrenar hace más de 3 días
          </Badge>
        )}

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Records personales</CardTitle>
          </CardHeader>
          <CardContent>
            {!records && !error && <p className="text-sm text-muted-foreground">Cargando...</p>}
            {records && records.length === 0 && (
              <p className="text-sm text-muted-foreground">
                Todavía no hay records. Aparecen cuando el cliente completa entrenamientos.
              </p>
            )}
            {records && records.length > 0 && (
              <div className="flex flex-col gap-6">
                {[...byExercise.entries()].map(([exerciseId, recs]) => (
                  <ExerciseRecords key={exerciseId} name={exerciseName(exerciseId)} records={recs} />
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border bg-background p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 text-sm font-medium">{value}</p>
    </div>
  );
}

function ExerciseRecords({ name, records }: { name: string; records: PersonalRecord[] }) {
  const maxWeight = records.find((r) => r.type === "max_weight");
  return (
    <div>
      <div className="mb-2 flex items-baseline justify-between">
        <span className="font-medium">{name}</span>
        {maxWeight && <span className="text-sm text-muted-foreground">{maxWeight.value} kg máx.</span>}
      </div>
      <RecordBars records={records} />
    </div>
  );
}

// Barras con CSS en vez de sumar una libreria de charts: son cuatro valores
// por ejercicio, no una serie temporal.
function RecordBars({ records }: { records: PersonalRecord[] }) {
  const max = Math.max(...records.map((r) => r.value), 1);
  return (
    <div className="flex flex-col gap-1.5">
      {records.map((r) => (
        <div key={r.type} className="flex items-center gap-3 text-sm">
          <span className="w-44 shrink-0 text-muted-foreground">{RECORD_LABEL[r.type] ?? r.type}</span>
          <div className="h-3 flex-1 overflow-hidden rounded-full bg-neutral-200 dark:bg-neutral-800">
            <div
              className="h-full rounded-full bg-[#C6FF4F]"
              style={{ width: `${Math.max((r.value / max) * 100, 2)}%` }}
            />
          </div>
          <span className="w-20 shrink-0 text-right font-mono text-xs tabular-nums">
            {r.value.toFixed(1)}
          </span>
        </div>
      ))}
    </div>
  );
}
