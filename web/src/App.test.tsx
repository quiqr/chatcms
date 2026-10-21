import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";

function mockFetch(response: Partial<Response> & { json: () => Promise<unknown> }) {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, status: 200, ...response }));
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("App", () => {
  it("renders the health result", async () => {
    mockFetch({
      json: async () => ({
        status: "ok",
        serverVersion: "0.1.0",
        schemaVersion: "1",
        database: "reachable",
      }),
    });

    render(<App />);

    expect(await screen.findByText("reachable")).toBeInTheDocument();
    expect(screen.getByText("0.1.0")).toBeInTheDocument();
    expect(screen.getByText("1")).toBeInTheDocument();
  });

  it("shows the server's message when the server is unhealthy", async () => {
    mockFetch({
      ok: false,
      status: 503,
      json: async () => ({
        error: { code: "unavailable", message: "The database is unreachable." },
      }),
    });

    render(<App />);

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "The database is unreachable.",
    );
  });

  it("shows a fallback message when the server answers with nothing useful", async () => {
    mockFetch({ ok: false, status: 500, json: async () => null });

    render(<App />);

    expect(await screen.findByRole("alert")).toHaveTextContent("status 500");
  });
});
