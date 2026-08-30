"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { clearTokens } from "@/lib/api";

const TABS = [
  { href: "/dashboard", label: "Clientes" },
  { href: "/programs", label: "Programas" },
  { href: "/suggestions", label: "Sugerencias IA" },
  { href: "/members", label: "Miembros" },
];

export function Nav() {
  const pathname = usePathname();
  const router = useRouter();

  function handleLogout() {
    clearTokens();
    router.replace("/login");
  }

  return (
    <header className="flex items-center justify-between border-b pb-4">
      <nav className="flex gap-1">
        {TABS.map((tab) => (
          <Link
            key={tab.href}
            href={tab.href}
            className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
              pathname === tab.href
                ? "bg-foreground text-background"
                : "text-muted-foreground hover:bg-neutral-200 dark:hover:bg-neutral-800"
            }`}
          >
            {tab.label}
          </Link>
        ))}
      </nav>
      <Button variant="outline" onClick={handleLogout}>
        Salir
      </Button>
    </header>
  );
}
