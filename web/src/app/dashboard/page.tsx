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
import { Nav } from "@/components/nav";
import { apiFetch, getAccessToken } from "@/lib/api";

type CoachClient = {
  client_user_id: string;
  full_name: string;
  avatar_url?: string;
  assignment_id?: string;
  assignment_name?: string;
  has_assignment: boolean;
  last_session_at?: string;
  streak_days: number;
  recent_prs: number;
  needs_attention: boolean;
};

function formatLastSession(iso?: string): string {
  if (!iso) return "Nunca entreno";
  const days = Math.floor((Date.now() - new Date(iso).getTime()) / 86_400_000);
  if (days <= 0) return "Hoy";
  if (days === 1) return "Ayer";
  return `Hace ${days} dias`;
}

export default function CoachDashboard() {
  const router = useRouter();
  const [clients, setClients] = useState<CoachClient[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!getAccessToken()) {
      router.replace("/login");
      return;
    }
    apiFetch("/v1/coach/clients")
      .then(async (res) => {
        if (!res.ok) throw new Error("No se pudo cargar la lista de clientes");
        setClients(await res.json());
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Error desconocido"));
  }, [router]);

  const needingAttention = clients?.filter((c) => c.needs_attention).length ?? 0;

  return (
    <div className="min-h-screen bg-neutral-50 p-6 dark:bg-neutral-950 sm:p-10">
      <div className="mx-auto flex max-w-4xl flex-col gap-6">
        <Nav />
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Mis clientes</h1>
          {clients && (
            <p className="text-sm text-muted-foreground">
              {needingAttention > 0
                ? `${needingAttention} necesitan atencion hoy`
                : "Todos al dia"}
            </p>
          )}
        </div>

        {error && <p className="text-sm text-red-500">{error}</p>}

        {!clients && !error && (
          <p className="text-sm text-muted-foreground">Cargando...</p>
        )}

        {clients && clients.length === 0 && (
          <p className="text-sm text-muted-foreground">
            Todavia no tienes clientes vinculados.
          </p>
        )}

        {clients && clients.length > 0 && (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Cliente</TableHead>
                <TableHead>Plan</TableHead>
                <TableHead>Ultimo entreno</TableHead>
                <TableHead>Racha</TableHead>
                <TableHead>PRs (14d)</TableHead>
                <TableHead>Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {clients.map((c) => (
                <TableRow key={c.client_user_id}>
                  <TableCell className="font-medium">
                    <a href={`/clients/${c.client_user_id}`} className="underline underline-offset-4">
                      {c.full_name}
                    </a>
                  </TableCell>
                  <TableCell>
                    {c.has_assignment ? c.assignment_name : "Sin plan asignado"}
                  </TableCell>
                  <TableCell>{formatLastSession(c.last_session_at)}</TableCell>
                  <TableCell>{c.streak_days} dias</TableCell>
                  <TableCell>{c.recent_prs}</TableCell>
                  <TableCell>
                    {c.needs_attention ? (
                      <Badge variant="destructive">Necesita atencion</Badge>
                    ) : (
                      <Badge variant="secondary">Al dia</Badge>
                    )}
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
