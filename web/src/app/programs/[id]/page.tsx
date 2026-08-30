"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Nav } from "@/components/nav";
import { apiFetch, getAccessToken } from "@/lib/api";

type Program = {
  id: string;
  name: string;
  description: string;
  goal: string;
  level: string;
  total_weeks: number;
  days_per_week: number;
  status: "draft" | "published" | "archived";
};

type Workout = {
  id: string;
  week_number: number;
  day_index: number;
  name: string;
};

type OrgMember = {
  user_id: string;
  full_name: string;
  role: string;
};

const GOALS = ["general_health", "hypertrophy", "strength", "fat_loss", "endurance"];
const LEVELS = ["beginner", "intermediate", "advanced"];

export default function ProgramEditorPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const [program, setProgram] = useState<Program | null>(null);
  const [workouts, setWorkouts] = useState<Workout[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const [form, setForm] = useState({ name: "", description: "", goal: "general_health", level: "beginner", total_weeks: 4, days_per_week: 3 });

  const [dayWeek, setDayWeek] = useState(1);
  const [dayIndex, setDayIndex] = useState(1);
  const [dayName, setDayName] = useState("");
  const [addingDay, setAddingDay] = useState(false);

  const [clients, setClients] = useState<OrgMember[] | null>(null);
  const [assignClientId, setAssignClientId] = useState("");
  const [assignDate, setAssignDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [assigning, setAssigning] = useState(false);

  function load() {
    apiFetch(`/v1/programs/${id}`)
      .then(async (res) => {
        if (!res.ok) throw new Error("No se pudo cargar el programa");
        const p: Program = await res.json();
        setProgram(p);
        setForm({
          name: p.name, description: p.description ?? "", goal: p.goal, level: p.level,
          total_weeks: p.total_weeks, days_per_week: p.days_per_week,
        });
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Error desconocido"));
    apiFetch(`/v1/programs/${id}/workouts`)
      .then(async (res) => {
        if (!res.ok) throw new Error("No se pudo cargar los días");
        setWorkouts(await res.json());
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Error desconocido"));
    apiFetch("/v1/org/members")
      .then((res) => res.json())
      .then((members: OrgMember[]) => setClients(members.filter((m) => m.role === "client")))
      .catch(() => setClients([]));
  }

  useEffect(() => {
    if (!getAccessToken()) {
      router.replace("/login");
      return;
    }
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, router]);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      const res = await apiFetch(`/v1/programs/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });
      if (!res.ok) throw new Error("No se pudo guardar el programa");
      setProgram(await res.json());
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error desconocido");
    } finally {
      setSaving(false);
    }
  }

  async function handleStatusChange(action: "publish" | "archive") {
    setError(null);
    try {
      const res = await apiFetch(`/v1/programs/${id}/${action}`, { method: "POST" });
      if (!res.ok) throw new Error("No se pudo cambiar el estado");
      setProgram(await res.json());
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error desconocido");
    }
  }

  async function handleAddDay(e: React.FormEvent) {
    e.preventDefault();
    setAddingDay(true);
    setError(null);
    try {
      const res = await apiFetch(`/v1/programs/${id}/workouts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ week_number: dayWeek, day_index: dayIndex, name: dayName }),
      });
      if (!res.ok) throw new Error("No se pudo agregar el día");
      const created: Workout = await res.json();
      setWorkouts((prev) => [...(prev ?? []), created]);
      setDayName("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error desconocido");
    } finally {
      setAddingDay(false);
    }
  }

  async function handleDeleteDay(workoutId: string) {
    setError(null);
    try {
      const res = await apiFetch(`/v1/workouts/${workoutId}`, { method: "DELETE" });
      if (!res.ok) throw new Error("No se pudo borrar el día");
      setWorkouts((prev) => prev?.filter((w) => w.id !== workoutId) ?? null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error desconocido");
    }
  }

  async function handleAssign(e: React.FormEvent) {
    e.preventDefault();
    setAssigning(true);
    setError(null);
    try {
      const res = await apiFetch(`/v1/programs/${id}/assign`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ client_user_id: assignClientId, start_date: assignDate }),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => null);
        throw new Error(body?.error ?? "No se pudo asignar el programa");
      }
      router.push("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error desconocido");
    } finally {
      setAssigning(false);
    }
  }

  if (!program) {
    return (
      <div className="min-h-screen bg-neutral-50 p-6 dark:bg-neutral-950 sm:p-10">
        <div className="mx-auto flex max-w-4xl flex-col gap-6">
          <Nav />
          {error ? <p className="text-sm text-red-500">{error}</p> : <p className="text-sm text-muted-foreground">Cargando...</p>}
        </div>
      </div>
    );
  }

  const weeks = Array.from({ length: form.total_weeks }, (_, i) => i + 1);
  const days = Array.from({ length: form.days_per_week }, (_, i) => i + 1);
  const workoutAt = (week: number, day: number) => workouts?.find((w) => w.week_number === week && w.day_index === day);

  return (
    <div className="min-h-screen bg-neutral-50 p-6 dark:bg-neutral-950 sm:p-10">
      <div className="mx-auto flex max-w-4xl flex-col gap-6">
        <Nav />

        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">{program.name}</h1>
            <Badge variant={program.status === "published" ? "default" : program.status === "archived" ? "secondary" : "outline"}>
              {program.status}
            </Badge>
          </div>
          <div className="flex gap-2">
            {program.status === "draft" && <Button onClick={() => handleStatusChange("publish")}>Publicar</Button>}
            {program.status === "published" && (
              <Button variant="outline" onClick={() => handleStatusChange("archive")}>
                Archivar
              </Button>
            )}
          </div>
        </div>

        {error && <p className="text-sm text-red-500">{error}</p>}

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Detalles</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSave} className="flex flex-col gap-4">
              <div className="grid gap-3 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-name">Nombre</Label>
                  <Input id="p-name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-goal">Objetivo</Label>
                  <select
                    id="p-goal"
                    value={form.goal}
                    onChange={(e) => setForm({ ...form, goal: e.target.value })}
                    className="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm"
                  >
                    {GOALS.map((g) => (
                      <option key={g} value={g}>{g}</option>
                    ))}
                  </select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-level">Nivel</Label>
                  <select
                    id="p-level"
                    value={form.level}
                    onChange={(e) => setForm({ ...form, level: e.target.value })}
                    className="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm"
                  >
                    {LEVELS.map((l) => (
                      <option key={l} value={l}>{l}</option>
                    ))}
                  </select>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="p-weeks">Semanas</Label>
                    <Input id="p-weeks" type="number" min={1} max={52} value={form.total_weeks} onChange={(e) => setForm({ ...form, total_weeks: Number(e.target.value) })} />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="p-days">Días/sem.</Label>
                    <Input id="p-days" type="number" min={1} max={7} value={form.days_per_week} onChange={(e) => setForm({ ...form, days_per_week: Number(e.target.value) })} />
                  </div>
                </div>
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="p-desc">Descripción</Label>
                <textarea
                  id="p-desc"
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  rows={2}
                  className="w-full rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
                />
              </div>
              <Button type="submit" disabled={saving} className="self-start">
                {saving ? "Guardando..." : "Guardar"}
              </Button>
            </form>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Días del programa</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <form onSubmit={handleAddDay} className="flex flex-wrap items-end gap-3">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="d-week">Semana</Label>
                <Input id="d-week" type="number" min={1} max={form.total_weeks} value={dayWeek} onChange={(e) => setDayWeek(Number(e.target.value))} className="w-20" />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="d-day">Día</Label>
                <Input id="d-day" type="number" min={1} max={7} value={dayIndex} onChange={(e) => setDayIndex(Number(e.target.value))} className="w-20" />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="d-name">Nombre</Label>
                <Input id="d-name" required value={dayName} onChange={(e) => setDayName(e.target.value)} placeholder="Empuje" className="w-48" />
              </div>
              <Button type="submit" disabled={addingDay}>{addingDay ? "Agregando..." : "Agregar día"}</Button>
            </form>

            <div className="overflow-x-auto">
              <table className="w-full border-collapse text-sm">
                <thead>
                  <tr>
                    <th className="p-2 text-left font-medium text-muted-foreground">Semana</th>
                    {days.map((d) => (
                      <th key={d} className="p-2 text-left font-medium text-muted-foreground">Día {d}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {weeks.map((w) => (
                    <tr key={w} className="border-t">
                      <td className="p-2 font-medium">{w}</td>
                      {days.map((d) => {
                        const workout = workoutAt(w, d);
                        return (
                          <td key={d} className="p-2">
                            {workout ? (
                              <div className="flex items-center gap-2">
                                <a href={`/programs/${id}/workouts/${workout.id}`} className="underline underline-offset-4">
                                  {workout.name}
                                </a>
                                <button
                                  type="button"
                                  onClick={() => handleDeleteDay(workout.id)}
                                  className="text-xs text-muted-foreground hover:text-red-500"
                                  aria-label={`Borrar ${workout.name}`}
                                >
                                  ✕
                                </button>
                              </div>
                            ) : (
                              <span className="text-muted-foreground">—</span>
                            )}
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>

        {program.status === "published" && (
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Asignar a un cliente</CardTitle>
            </CardHeader>
            <CardContent>
              {clients && clients.length === 0 && (
                <p className="text-sm text-muted-foreground">No hay clientes en el gimnasio todavía.</p>
              )}
              {clients && clients.length > 0 && (
                <form onSubmit={handleAssign} className="flex flex-wrap items-end gap-3">
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="a-client">Cliente</Label>
                    <select
                      id="a-client"
                      required
                      value={assignClientId}
                      onChange={(e) => setAssignClientId(e.target.value)}
                      className="h-8 w-56 rounded-lg border border-input bg-transparent px-2.5 text-sm"
                    >
                      <option value="" disabled>Elegir cliente</option>
                      {clients.map((c) => (
                        <option key={c.user_id} value={c.user_id}>{c.full_name}</option>
                      ))}
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="a-date">Fecha de inicio</Label>
                    <Input id="a-date" type="date" value={assignDate} onChange={(e) => setAssignDate(e.target.value)} />
                  </div>
                  <Button type="submit" disabled={assigning}>{assigning ? "Asignando..." : "Asignar"}</Button>
                </form>
              )}
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
