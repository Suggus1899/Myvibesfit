const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const ACCESS_KEY = "myvibesfit_access_token";
const REFRESH_KEY = "myvibesfit_refresh_token";

export function getAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(ACCESS_KEY);
}

export function setTokens(access: string, refresh: string) {
  localStorage.setItem(ACCESS_KEY, access);
  localStorage.setItem(REFRESH_KEY, refresh);
}

export function clearTokens() {
  localStorage.removeItem(ACCESS_KEY);
  localStorage.removeItem(REFRESH_KEY);
}

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

export type LoginResponse = {
  access_token: string;
  refresh_token: string;
  user: { id: string; email: string; full_name: string; org_id?: string | null; role?: string };
};

export async function login(email: string, password: string): Promise<LoginResponse> {
  const res = await fetch(`${API_URL}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) throw new ApiError(res.status, "Credenciales invalidas");
  return res.json();
}

export async function register(email: string, password: string, fullName: string): Promise<LoginResponse> {
  const res = await fetch(`${API_URL}/v1/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password, full_name: fullName }),
  });
  if (!res.ok) throw new ApiError(res.status, "No se pudo crear la cuenta");
  return res.json();
}

async function refresh(): Promise<string> {
  const refreshToken = typeof window !== "undefined" ? localStorage.getItem(REFRESH_KEY) : null;
  if (!refreshToken) throw new ApiError(401, "Sin sesion");

  const res = await fetch(`${API_URL}/v1/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
  if (!res.ok) throw new ApiError(res.status, "Sesion expirada");
  const data: LoginResponse = await res.json();
  setTokens(data.access_token, data.refresh_token);
  return data.access_token;
}

// apiFetch reintenta una vez con refresh si el access token vencio (401) —
// el token dura 15 min y el coach no deberia perder la sesion cada rato.
export async function apiFetch(path: string, init?: RequestInit): Promise<Response> {
  let token = getAccessToken();
  if (!token) throw new ApiError(401, "Sin sesion");

  let res = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: { ...init?.headers, Authorization: `Bearer ${token}` },
  });

  if (res.status === 401) {
    token = await refresh();
    res = await fetch(`${API_URL}${path}`, {
      ...init,
      headers: { ...init?.headers, Authorization: `Bearer ${token}` },
    });
  }

  return res;
}
