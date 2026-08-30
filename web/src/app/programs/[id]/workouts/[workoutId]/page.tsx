"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Nav } from "@/components/nav";
import { apiFetch, getAccessToken } from "@/lib/api";

type Exercise = { id: string; name: string; primary_muscle: string };

type ProgramExercise = {
  id: string;
  exercise_id: string;
  order_index: number;
  target_sets: number;
  target_reps_min?: number;
  target_reps_max?: number;
  target_rpe?: number;
  rest_seconds: number;
};

export default function WorkoutEditorPage() {
  const { id, workoutId } = useParams<{ id: string; workoutId: string }>();
  const router = useRouter();

  const [workoutName, setWorkoutName] = useState<string | null>(null);
  const [exercises, setExercises] = useState<ProgramExercise[] | null>(null);
  const [catalog, setCatalog] = useState<Exercise[]>([]);
  const [error, setError] = useState<string | null>(null);

  const [form, setForm] = useState({ exerciseId: "", sets: 3, repsMin: 8, repsMax: 12, rpe: "", rest: 90 });
  const [saving, setSaving] = useState(false);

  function load() {
    apiFetch(`/v1/programs/${id}/workouts`)
      .then((res) => res.json())
      .then((list: { id: string; name: string }[]) => {
        setWorkoutName(list.find((w) => w.id === workoutId)?.name ?? "Día");
      });
    apiFetch(`/v1/workouts/${workoutId}/exercises`)
      .then(async (res) => {
        if (!res.ok) throw new Error("No se pudo cargar los ejercicios");
        setExercises(await res.json());
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Error desconocido"));
    apiFetch("/v1/exercises?limit=100")
      .then((res) => res.json())
      .then(setCatalog)
      .catch(() => setCatalog([]));
  }

  useEffect(() => {
    if (!getAccessToken()) {
      router.replace("/login");
      return;
    }
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, workoutId, router]);

  function exerciseName(exerciseId: string) {
    return catalog.find((e) => e.id === exerciseId)?.name ?? exerciseId.slice(0, 8);
  }

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      const res = await apiFetch(`/v1/workouts/${workoutId}/exercises`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          exercise_id: form.exerciseId,
          order_index: (exercises?.length ?? 0) + 1,
          target_sets: form.sets,
          target_reps_min: form.repsMin || null,
          target_reps_max: form.repsMax || null,
          target_rpe: form.rpe ? Number(form.rpe) : null,
          rest_seconds: form.rest,
        }),
      });
      if (!res.ok) throw new Error("No se pudo agregar el ejercicio");
      const created: ProgramExercise = await res.json();
      setExercises((prev) => [...(prev ?? []), created]);
      setForm({ ...form, exerciseId: "" });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error desconocido");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(exerciseRowId: string) {
    setError(null);
    try {
      const res = await apiFetch(`/v1/program-exercises/${exerciseRowId}`, { method: "DELETE" });
      if (!res.ok) throw new Error("No se pudo borrar el ejercicio");
      setExercises((prev) => prev?.filter((e) => e.id !== exerciseRowId) ?? null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error desconocido");
    }
  }

  return (
    <div className="min-h-screen bg-neutral-50 p-6 dark:bg-neutral-950 sm:p-10">
      <div className="mx-auto flex max-w-3xl flex-col gap-6">
        <Nav />
        <div>
          <a href={`/programs/${id}`} className="text-sm text-muted-foreground underline underline-offset-4">
            ← Volver al programa
          </a>
          <h1 className="mt-1 text-2xl font-bold tracking-tight">{workoutName ?? "Día"}</h1>
        </div>

        {error && <p className="text-sm text-red-500">{error}</p>}

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Agregar ejercicio</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleAdd} className="flex flex-wrap items-end gap-3">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="ex">Ejercicio</Label>
                <select
                  id="ex"
                  required
                  value={form.exerciseId}
                  onChange={(e) => setForm({ ...form, exerciseId: e.target.value })}
                  className="h-8 w-56 rounded-lg border border-input bg-transparent px-2.5 text-sm"
                >
                  <option value="" disabled>Elegir ejercicio</option>
                  {catalog.map((ex) => (
                    <option key={ex.id} value={ex.id}>{ex.name}</option>
                  ))}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="sets">Series</Label>
                <Input id="sets" type="number" min={1} value={form.sets} onChange={(e) => setForm({ ...form, sets: Number(e.target.value) })} className="w-16" />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="reps-min">Reps min</Label>
                <Input id="reps-min" type="number" min={0} value={form.repsMin} onChange={(e) => setForm({ ...form, repsMin: Number(e.target.value) })} className="w-16" />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="reps-max">Reps max</Label>
                <Input id="reps-max" type="number" min={0} value={form.repsMax} onChange={(e) => setForm({ ...form, repsMax: Number(e.target.value) })} className="w-16" />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="rpe">RPE</Label>
                <Input id="rpe" type="number" min={1} max={10} step={0.5} value={form.rpe} onChange={(e) => setForm({ ...form, rpe: e.target.value })} className="w-16" />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="rest">Descanso (s)</Label>
                <Input id="rest" type="number" min={0} value={form.rest} onChange={(e) => setForm({ ...form, rest: Number(e.target.value) })} className="w-20" />
              </div>
              <Button type="submit" disabled={saving}>{saving ? "Agregando..." : "Agregar"}</Button>
            </form>
          </CardContent>
        </Card>

        {!exercises && !error && <p className="text-sm text-muted-foreground">Cargando...</p>}
        {exercises && exercises.length === 0 && (
          <p className="text-sm text-muted-foreground">Todavía no hay ejercicios en este día.</p>
        )}
        {exercises && exercises.length > 0 && (
          <div className="flex flex-col gap-2">
            {[...exercises]
              .sort((a, b) => a.order_index - b.order_index)
              .map((ex) => (
                <div key={ex.id} className="flex items-center justify-between rounded-lg border p-3 text-sm">
                  <div>
                    <span className="font-medium">{exerciseName(ex.exercise_id)}</span>
                    <span className="ml-2 text-muted-foreground">
                      {ex.target_sets}×{ex.target_reps_min ?? "?"}
                      {ex.target_reps_max ? `-${ex.target_reps_max}` : ""}
                      {ex.target_rpe ? ` @RPE ${ex.target_rpe}` : ""} · {ex.rest_seconds}s desc.
                    </span>
                  </div>
                  <button
                    type="button"
                    onClick={() => handleDelete(ex.id)}
                    className="text-xs text-muted-foreground hover:text-red-500"
                  >
                    Quitar
                  </button>
                </div>
              ))}
          </div>
        )}
      </div>
    </div>
  );
}
