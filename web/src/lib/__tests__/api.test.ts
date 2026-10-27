import { afterEach, describe, expect, it, vi } from "vitest";

import { get, post, RequestFailed, whatWentWrong } from "../api";
import { speak } from "../say";

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

describe("what a screen says went wrong", () => {
  it("says a refusal the server named, in the language in force", () => {
    speak("nl");
    const refused = new RequestFailed({
      code: "refused",
      message: "That address is already taken.",
      key: "refused.address-already-taken",
    });

    expect(whatWentWrong(refused)).toBe("Dat adres is al bezet.");
    speak("en");
  });

  it("fills in what the refusal carried with it", () => {
    const refused = new RequestFailed({
      code: "refused",
      message: "sam@example.org already has an account.",
      key: "refused.address-has-an-account",
      detail: { address: "sam@example.org" },
    });

    expect(whatWentWrong(refused)).toBe("sam@example.org already has an account.");
  });

  it("falls back to the server's own words for a refusal nobody translated", () => {
    const refused = new RequestFailed({
      code: "refused",
      message: "Something nobody has written a sentence for yet.",
      key: "refused.nobody-has-this-key",
    });

    expect(whatWentWrong(refused)).toBe("Something nobody has written a sentence for yet.");
  });

  it("says its own sentence for anything the browser raised", () => {
    // A failed connection is in no language anybody chose.
    expect(whatWentWrong(new Error("Failed to fetch"))).toBe("That did not work.");
    expect(whatWentWrong(new Error("Failed to fetch"), "common.cannot-reach-the-server")).toBe(
      "We could not reach the server. Nothing is lost; please try again.",
    );
  });
});
