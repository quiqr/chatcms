export type Health = {
  status: string;
  serverVersion: string;
  schemaVersion: string;
  database: string;
};

export type ApiError = {
  error: {
    code: string;
    message: string;
  };
};

export async function fetchHealth(): Promise<Health> {
  const response = await fetch("/api/health", {
    headers: { Accept: "application/json" },
  });

  const body: unknown = await response.json().catch(() => null);

  if (!response.ok) {
    throw new Error(messageFrom(body, response.status));
  }
  return body as Health;
}

function messageFrom(body: unknown, status: number): string {
  if (
    typeof body === "object" &&
    body !== null &&
    "error" in body &&
    typeof (body as ApiError).error?.message === "string"
  ) {
    return (body as ApiError).error.message;
  }
  return `The server answered with status ${status}.`;
}
