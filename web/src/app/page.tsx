import type { Metadata } from "next";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export const metadata: Metadata = {
  title: "Myvibesfit — Software para gimnasios y coaches",
  description:
    "Programas de entrenamiento, seguimiento de progreso y gamificacion para que tus clientes entrenen mas y vos gestiones menos.",
};

const FEATURES = [
  {
    title: "Programas y asignaciones",
    description:
      "Armá plantillas de entrenamiento reutilizables y asignalas a cada cliente en segundos.",
    icon: (
      <path d="M6 4v16M18 4v16M6 8h4M14 8h4M6 12h12M6 16h4M14 16h4" />
    ),
  },
  {
    title: "Progreso automático",
    description:
      "Cada set sincronizado detecta records personales y arma el historial de fuerza sin trabajo manual.",
    icon: <path d="M4 18l5-6 4 4 7-9M14 6h6v6" />,
  },
  {
    title: "Rachas, XP y logros",
    description:
      "Tus clientes entrenan más seguido cuando el progreso se siente como un juego, no como una obligación.",
    icon: <path d="M12 3c3 3 5 6 5 9a5 5 0 0 1-10 0c0-1.3.6-2.4 1.5-3.5.3 1 1 1.5 1.5 1 .3-1.5-.5-3-.5-5 1 .5 2 1.7 2.5 3.5" />,
  },
  {
    title: "Sugerencias con IA",
    description:
      "El sistema propone ajustes de carga y volumen a partir del historial real de cada cliente; vos revisás y aprobás.",
    icon: <path d="M12 3v3M12 18v3M4.2 4.2l2.1 2.1M17.7 17.7l2.1 2.1M3 12h3M18 12h3M4.2 19.8l2.1-2.1M17.7 6.3l2.1-2.1M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8Z" />,
  },
];

export default function LandingPage() {
  return (
    <div className="flex min-h-screen flex-col bg-background">
      <header className="border-b">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <span className="text-lg font-bold tracking-tight">
            Myvibes<span className="text-[#5c7a00]">fit</span>
          </span>
          <nav className="flex items-center gap-2">
            <Link href="/login">
              <Button variant="outline">Iniciar sesión</Button>
            </Link>
          </nav>
        </div>
      </header>

      <main className="flex-1">
        <section className="mx-auto grid max-w-5xl items-center gap-10 px-6 py-20 sm:py-28 md:grid-cols-2">
          <div className="flex flex-col gap-6">
            <h1 className="text-4xl font-bold tracking-tight sm:text-5xl">
              La app de entrenamiento para tu gimnasio
            </h1>
            <p className="text-lg text-muted-foreground">
              Programas, seguimiento de progreso y gamificación en un solo lugar.
              Tus clientes entrenan más, vos gestionás menos.
            </p>
            <div className="flex flex-wrap gap-3">
              <Link href="/login">
                <Button size="lg" className="bg-[#C6FF4F] text-neutral-900 hover:bg-[#b3e844]">
                  Empezar ahora
                </Button>
              </Link>
              <a href="#features">
                <Button size="lg" variant="ghost">
                  Ver características
                </Button>
              </a>
            </div>
          </div>

          <HeroIllustration />
        </section>

        <section id="features" className="border-t bg-neutral-50 py-20 dark:bg-neutral-950">
          <div className="mx-auto max-w-5xl px-6">
            <div className="mb-12 text-center">
              <h2 className="text-3xl font-bold tracking-tight">Todo lo que necesita tu equipo</h2>
              <p className="mt-3 text-muted-foreground">
                Pensado para coaches que quieren pasar menos tiempo armando planillas.
              </p>
            </div>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              {FEATURES.map((f) => (
                <Card key={f.title}>
                  <CardHeader>
                    <div className="mb-2 flex size-10 items-center justify-center rounded-lg bg-[#C6FF4F]/20 text-[#5c7a00]">
                      <svg
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="1.8"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        className="size-5"
                      >
                        {f.icon}
                      </svg>
                    </div>
                    <CardTitle>{f.title}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm text-muted-foreground">{f.description}</p>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </section>

        <section className="border-t py-20">
          <div className="mx-auto flex max-w-3xl flex-col items-center gap-5 px-6 text-center">
            <h2 className="text-3xl font-bold tracking-tight">
              Llevá tu gimnasio a Myvibesfit
            </h2>
            <p className="text-muted-foreground">
              Creá tu cuenta de coach y empezá a asignar programas hoy mismo.
            </p>
            <Link href="/login">
              <Button size="lg" className="bg-[#C6FF4F] text-neutral-900 hover:bg-[#b3e844]">
                Empezar ahora
              </Button>
            </Link>
          </div>
        </section>
      </main>

      <footer className="border-t py-8">
        <div className="mx-auto max-w-5xl px-6 text-sm text-muted-foreground">
          © {new Date().getFullYear()} Myvibesfit
        </div>
      </footer>
    </div>
  );
}

function HeroIllustration() {
  return (
    <svg
      viewBox="0 0 420 340"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="w-full max-w-md justify-self-center"
      role="img"
      aria-label="Panel de progreso de entrenamiento con racha y barra de carga"
    >
      <rect x="20" y="20" width="380" height="300" rx="24" className="fill-neutral-100 dark:fill-neutral-900" />
      <rect x="48" y="52" width="180" height="16" rx="8" className="fill-neutral-300 dark:fill-neutral-700" />
      <rect x="48" y="80" width="120" height="12" rx="6" className="fill-neutral-200 dark:fill-neutral-800" />

      <rect x="48" y="120" width="324" height="90" rx="16" fill="#C6FF4F" />
      <path
        d="M70 190 L110 150 L145 175 L190 120 L230 160 L270 135 L310 165 L350 140"
        stroke="#3d5200"
        strokeWidth="6"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
      />

      <g>
        <circle cx="90" cy="260" r="34" className="fill-neutral-100 stroke-[#C6FF4F] dark:fill-neutral-900" strokeWidth="8" />
        <text x="90" y="267" textAnchor="middle" className="fill-neutral-900 dark:fill-neutral-50" fontSize="22" fontWeight="700">
          12
        </text>
      </g>
      <rect x="140" y="242" width="90" height="14" rx="7" className="fill-neutral-300 dark:fill-neutral-700" />
      <rect x="140" y="264" width="140" height="12" rx="6" className="fill-neutral-200 dark:fill-neutral-800" />

      <rect x="330" y="238" width="42" height="42" rx="12" fill="#C6FF4F" />
      <path
        d="M351 250c4 4 6.5 8 6.5 11.5a6.5 6.5 0 0 1-13 0c0-1.7.8-3.1 2-4.6.4 1.3 1.3 2 2 1.3.4-2-.7-3.9-.7-6.5 1.3.6 2.6 2.2 3.2 4.6"
        stroke="#3d5200"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
      />
    </svg>
  );
}
