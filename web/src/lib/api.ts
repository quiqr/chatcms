const REQUEST_TOKEN_COOKIE = "chatcms_request_token";
const REQUEST_TOKEN_HEADER = "X-Chatcms-Request-Token";

export type ApiError = {
  code: string;
  message: string;
};

export class RequestFailed extends Error {
  readonly code: string;

  constructor(error: ApiError) {
    super(error.message);
    this.code = error.code;
  }
}

function requestToken() {
  const match = document.cookie.match(new RegExp(`(?:^|; )${REQUEST_TOKEN_COOKIE}=([^;]*)`));
  return match ? decodeURIComponent(match[1]!) : "";
}

export async function get<T>(path: string): Promise<T> {
  return send<T>("GET", path);
}

export async function post<T>(path: string, body?: unknown): Promise<T> {
  return send<T>("POST", path, body);
}

export async function put<T>(path: string, body?: unknown): Promise<T> {
  return send<T>("PUT", path, body);
}

export async function remove<T>(path: string): Promise<T> {
  return send<T>("DELETE", path);
}

/**
 * ensureRequestToken returns the token this page must repeat in a header, and
 * asks the server for one first if the page does not have it. The server sets
 * the cookie on any request it answers, so a page it served itself always has
 * one; a page served by something else, such as the development server, or a
 * page whose cookie was cleared while it stayed open, does not.
 */
async function ensureRequestToken(): Promise<string> {
  const held = requestToken();
  if (held) {
    return held;
  }

  try {
    await fetch("/api/health", { credentials: "same-origin" });
  } catch {
    // The request that follows will fail with something a person can read.
  }
  return requestToken();
}

/**
 * sendForm sends a file the way a browser sends one, which is the only way to
 * send a picture without turning it into text first.
 */
export async function sendForm<T>(path: string, body: FormData): Promise<T> {
  const response = await fetch(path, {
    method: "POST",
    headers: {
      Accept: "application/json",
      [REQUEST_TOKEN_HEADER]: await ensureRequestToken(),
    },
    credentials: "same-origin",
    body,
  });

  const payload: unknown = await response.json().catch(() => null);
  if (!response.ok) {
    askToSignIn(response.status, path);
    throw new RequestFailed(errorFrom(payload, response.status));
  }
  return payload as T;
}

/**
 * askToSignIn sends somebody whose session has gone to the sign-in screen,
 * remembering where they were so that signing in puts them back. It lives
 * here rather than on one screen because every screen can be the one open
 * when a session runs out, and a screen that answers "nothing here" to that
 * is a screen that looks broken.
 *
 * Signing in itself is left alone: a refused password is a 401 too, and it
 * belongs on the form rather than in a reload.
 */
function askToSignIn(status: number, path: string) {
  if (status !== 401) return;
  if (typeof window === "undefined") return;
  if (path.startsWith("/api/sign-in") || path.startsWith("/api/set-password")) return;

  const here = window.location.pathname;
  if (here === "/sign-in") return;

  const back = here + window.location.search;
  window.location.assign(`/sign-in?then=${encodeURIComponent(back)}`);
}

async function send<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { Accept: "application/json" };

  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }
  if (method !== "GET" && method !== "HEAD") {
    headers[REQUEST_TOKEN_HEADER] = await ensureRequestToken();
  }

  const response = await fetch(path, {
    method,
    headers,
    credentials: "same-origin",
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const payload: unknown = await response.json().catch(() => null);

  if (!response.ok) {
    askToSignIn(response.status, path);
    throw new RequestFailed(errorFrom(payload, response.status));
  }
  return payload as T;
}

function errorFrom(payload: unknown, status: number): ApiError {
  if (
    typeof payload === "object" &&
    payload !== null &&
    "error" in payload &&
    typeof (payload as { error?: ApiError }).error?.message === "string"
  ) {
    return (payload as { error: ApiError }).error;
  }
  return { code: "unknown", message: `The server answered with status ${status}.` };
}
