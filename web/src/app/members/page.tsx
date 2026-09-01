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

type OrgMember = {
  membership_id: string;
  user_id: string;
  full_name: string;
  email: string;
  role: string;
  status: string;
};

type Org = {
  id: string;
  name: string;
  join_code: string;
};

const ROLE_OPTIONS = ["client", "coach", "admin"];

export default function MembersPage() {
  const router = useRouter();
  const [org, setOrg] = useState<Org | null>(null);
  const [members, setMembers] = useState<OrgMember[] | null>(null);
  const [myRole, setMyRole] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [savingId, setSavingId] = useState<string | null>(null);

  function load() {
    apiFetch("/v1/me")
      .then((res) => res.json())
      .then((me) => setMyRole(me.role ?? null));
    apiFetch("/v1/org")
      .then(async (res) => {
        if (!res.ok) throw new Error("No se pudo cargar el gimnasio");
        setOrg(await res.json());
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Error desconocido"));
    apiFetch("/v1/org/members")
      .then(async (res) => {
        if (!res.ok) throw new Error("No se pudo cargar los miembros");
        setMembers(await res.json());
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Error desconocido"));
  }

  useEffect(() => {
    if (!getAccessToken()) {
      router.replace("/login");
      return;
    }
    load();
     
  }, [router]);

  async function changeRole(membershipId: string, role: string) {
    setSavingId(membershipId);
    try {
      const res = await apiFetch(`/v1/org/members/${membershipId}/role`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ role }),
      });
      if (!res.ok) throw new Error("No se pudo cambiar el rol");
      setMembers((prev) =>
        prev ? prev.map((m) => (m.membership_id === membershipId ? { ...m, role } : m)) : prev
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error desconocido");
    } finally {
      setSavingId(null);
    }
  }

  const canEditRoles = myRole === "owner";

  return (
    <div className="min-h-screen bg-neutral-50 p-6 dark:bg-neutral-950 sm:p-10">
      <div className="mx-auto flex max-w-4xl flex-col gap-6">
        <Nav />
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Miembros</h1>
          {org && (
            <p className="text-sm text-muted-foreground">
              {org.name} — código de acceso{" "}
              <span className="rounded bg-neutral-200 px-1.5 py-0.5 font-mono text-xs dark:bg-neutral-800">
                {org.join_code}
              </span>{" "}
              (compartilo con tus clientes para que se unan desde la app)
            </p>
          )}
        </div>

        {error && <p className="text-sm text-red-500">{error}</p>}
        {!members && !error && <p className="text-sm text-muted-foreground">Cargando...</p>}

        {members && members.length > 0 && (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nombre</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Rol</TableHead>
                <TableHead>Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {members.map((m) => (
                <TableRow key={m.membership_id}>
                  <TableCell className="font-medium">{m.full_name}</TableCell>
                  <TableCell>{m.email}</TableCell>
                  <TableCell>
                    {canEditRoles && m.role !== "owner" ? (
                      <select
                        value={m.role}
                        disabled={savingId === m.membership_id}
                        onChange={(e) => changeRole(m.membership_id, e.target.value)}
                        className="h-8 rounded-lg border border-input bg-transparent px-2 text-sm"
                      >
                        {ROLE_OPTIONS.map((r) => (
                          <option key={r} value={r}>
                            {r}
                          </option>
                        ))}
                      </select>
                    ) : (
                      <Badge variant="secondary">{m.role}</Badge>
                    )}
                  </TableCell>
                  <TableCell>{m.status}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </div>
    </div>
  );
}
