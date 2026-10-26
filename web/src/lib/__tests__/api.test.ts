import { afterEach, describe, expect, it, vi } from "vitest";

import { get, post } from "../api";

function answers(status: number, body: unknown) {
  return vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  });
}

const refused = { error: { code: "unauthorized", message: "Please sign in first." } };

function watchTheAddressBar() {
  const assign = vi.fn();
  vi.spyOn(window, "location", "get").mockReturnValue({
    ...window.location,
    pathname: "/p/korenaar/content",
    search: "",
    assign,
  } as unknown as Location);
  return assign;
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  document.cookie = "chatcms_request_token=; max-age=0; path=/";
});

describe("a session that has run out", () => {
  it("sends somebody to sign in, and remembers where they were", async () => {
    vi.stubGlobal("fetch", answers(401, refused));
    const assign = watchTheAddressBar();

    await expect(get("/api/organisations")).rejects.toThrow("Please sign in first.");

    expect(assign).toHaveBeenCalledWith("/sign-in?then=%2Fp%2Fkorenaar%2Fcontent");
  });

  it("leaves a refused password on the sign-in form", async () => {
    document.cookie = "chatcms_request_token=a-token; path=/";
    vi.stubGlobal("fetch", answers(401, refused));

    const assign = vi.fn();
    vi.spyOn(window, "location", "get").mockReturnValue({
      ...window.location,
      pathname: "/sign-in",
      search: "",
      assign,
    } as unknown as Location);

    await expect(post("/api/sign-in", { email: "a@b.c", password: "wrong" })).rejects.toThrow();

    expect(assign).not.toHaveBeenCalled();
  });

  it("does not send somebody who is already there", async () => {
    vi.stubGlobal("fetch", answers(401, refused));

    const assign = vi.fn();
    vi.spyOn(window, "location", "get").mockReturnValue({
      ...window.location,
      pathname: "/sign-in",
      search: "",
      assign,
    } as unknown as Location);

    await expect(get("/api/me")).rejects.toThrow();

    expect(assign).not.toHaveBeenCalled();
  });

  it("leaves other failures where they are", async () => {
    vi.stubGlobal("fetch", answers(404, { error: { code: "not-found", message: "Nothing here." } }));
    const assign = watchTheAddressBar();

    await expect(get("/api/organisations/nope/projects")).rejects.toThrow("Nothing here.");

    expect(assign).not.toHaveBeenCalled();
  });
});
