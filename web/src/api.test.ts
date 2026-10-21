import { afterEach, describe, expect, it, vi } from "vitest";

import { fetchHealth } from "./api";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("fetchHealth", () => {
  it("requests the API path the server serves in both modes", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ status: "ok" }),
    });
    vi.stubGlobal("fetch", fetchMock);

    await fetchHealth();

    expect(fetchMock).toHaveBeenCalledWith("/api/health", expect.anything());
  });

  it("throws the server's message when the request fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 503,
        json: async () => ({ error: { code: "unavailable", message: "Not today." } }),
      }),
    );

    await expect(fetchHealth()).rejects.toThrow("Not today.");
  });

  it("throws a readable message when the body is not JSON", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 502,
        json: async () => {
          throw new Error("not json");
        },
      }),
    );

    await expect(fetchHealth()).rejects.toThrow("status 502");
  });
});
