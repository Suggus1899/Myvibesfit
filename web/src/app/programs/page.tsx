"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Nav } from "@/components/nav";
import { apiFetch, getAccessToken } from "@/lib/api";

type Program = {
  id: string;
  name: string;
  level: string;
  goal: string;
  total_weeks: number;
  days_per_week: number;
  status: "draft" | "published" | "archived";
};

const STATUS_LABEL: Record<Program["status"], string> = {
  draft: "Borrador",
  published: "Publicado",
  archived: "Archivado",
};

const STATUS_VARIANT: Record<Program["status"], "secondary" | "default" | "outline"> = {
  draft: "outline",
  published: "default",
  archived: "secondary",
};

export default function ProgramsPage() {
  const router = useRouter();
  const [programs, setPrograms] = useState<Program[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [name, setName] = useState("");
  const [totalWeeks, setTotalWeeks] = useState(4);
  const [daysPerWeek, setDaysPerWeek] = useState(3);
  const [creating, setCreating] = useState(false);

  function load() {
    apiFetch("/v1/programs")
      .then(async (res) => {
        if (!res.ok) throw new Error("No se pudo cargar los programas");
        setPrograms(await res.json());
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

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setCreating(true);
    setError(null);
    try {
      const res = await apiFetch("/v1/programs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, total_weeks: totalWeeks, days_per_week: daysPerWeek }),
      });
      if (!res.ok) throw new Error("No se pudo crear el programa");
      const created: Program = await res.json();
      router.push(`/programs/${created.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error desconocido");
      setCreating(false);
    }
  }

  return (
    <div className="min-h-screen bg-neutral-50 p-6 dark:bg-neutral-950 sm:p-10">
      <div className="mx-auto flex max-w-4xl flex-col gap-6">
        <Nav />
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Programas</h1>
          <p className="text-sm text-muted-foreground">Plantillas de entrenamiento reutilizables.</p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Nuevo programa</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleCreate} className="flex flex-wrap items-end gap-3">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="name">Nombre</Label>
                <Input
                  id="name"
                  required
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Plan Fuerza 4 días"
                  className="w-56"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="weeks">Semanas</Label>
                <Input
                  id="weeks"
                  type="number"
                  min={1}
                  max={52}
                  value={totalWeeks}
                  onChange={(e) => setTotalWeeks(Number(e.target.value))}
                  className="w-20"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="days">Días/semana</Label>
                <Input
                  id="days"
                  type="number"
                  min={1}
                  max={7}
                  value={daysPerWeek}
                  onChange={(e) => setDaysPerWeek(Number(e.target.value))}
                  className="w-20"
                />
              </div>
              <Button type="submit" disabled={creating}>
                {creating ? "Creando..." : "Crear"}
              </Button>
            </form>
          </CardContent>
        </Card>

        {error && <p className="text-sm text-red-500">{error}</p>}
        {!programs && !error && <p className="text-sm text-muted-foreground">Cargando...</p>}
        {programs && programs.length === 0 && (
          <p className="text-sm text-muted-foreground">Todavía no creaste ningún programa.</p>
        )}

        {programs && programs.length > 0 && (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nombre</TableHead>
                <TableHead>Semanas</TableHead>
                <TableHead>Días/sem.</TableHead>
                <TableHead>Estado</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {programs.map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="font-medium">{p.name}</TableCell>
                  <TableCell>{p.total_weeks}</TableCell>
                  <TableCell>{p.days_per_week}</TableCell>
                  <TableCell>
                    <Badge variant={STATUS_VARIANT[p.status]}>{STATUS_LABEL[p.status]}</Badge>
                  </TableCell>
                  <TableCell>
                    <a href={`/programs/${p.id}`} className="text-sm font-medium underline underline-offset-4">
                      Abrir
                    </a>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </div>
    </div>
  );
}
