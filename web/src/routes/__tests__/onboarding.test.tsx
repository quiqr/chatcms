import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";

import { Inside } from "./here";import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { OnboardingPage } from "../OnboardingPage";

const base = "/api/organisations/technative/projects/korenaar";

const aReadySite = {
  kind: "repository",
  remote: "git@example.org:korenaar/site.git",
  isHugoSite: true,
  hasModel: true,
  ready: true,
  pages: 3,
  collections: 2,
};

function server(overrides: Record<string, unknown> = {}) {
  const answers: Record<string, unknown> = {
    [`${base}/site`]: { kind: "none", isHugoSite: false, hasModel: false, ready: false, pages: 0, collections: 0 },
    [`${base}/key`]: { publicKey: "", advice: "This project has no key yet. Make one to continue." },
    [`${base}/publishing`]: { to: "main-line", form: "source" },
    ...overrides,
  };

  return vi.fn().mockImplementation((path: string) =>
    Promise.resolve({
      ok: true,
      status: 200,
      json: async () => answers[path] ?? {},
    }),
  );
}

function renderPage() {
  return render(
    <Inside><MemoryRouter initialEntries={["/o/technative/p/korenaar/setup"]}>
      <Routes>
        <Route path="/o/:organisation/p/:project/setup" element={<OnboardingPage />} />
      </Routes>
    </MemoryRouter></Inside>,
  );
}

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("the onboarding screen", () => {
  it("shows the steps the designs draw", async () => {
    vi.stubGlobal("fetch", server());
    renderPage();

    const steps = await screen.findByRole("navigation", { name: "Setting up this project" });
    for (const label of [
      "Where the site comes from",
      "Key for the git host",
      "How it is published",
      "Invite people",
    ]) {
      expect(within(steps).getByText(label)).toBeInTheDocument();
    }
  });

  it("offers the three ways a site can arrive", async () => {
    vi.stubGlobal("fetch", server());
    renderPage();

    const choice = await screen.findByRole("radiogroup", { name: "Where the site comes from" });
    for (const label of ["A site already in git", "From Quiqr Desktop", "Start from a template"]) {
      expect(within(choice).getByRole("radio", { name: label })).toBeInTheDocument();
    }
  });

  it("brings a site in and reports what arrived", async () => {
    const fetchMock = vi.fn().mockImplementation((path: string, options?: { method?: string }) => {
      if (options?.method === "POST" && path === `${base}/import`) {
        return Promise.resolve({ ok: true, status: 201, json: async () => aReadySite });
      }
      if (path === `${base}/site`) {
        return Promise.resolve({ ok: true, status: 200, json: async () => aReadySite });
      }
      return Promise.resolve({ ok: true, status: 200, json: async () => ({ to: "main-line", form: "source", publicKey: "", advice: "" }) });
    });
    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    renderPage();

    await screen.findByRole("radiogroup", { name: "Where the site comes from" });
    await user.type(screen.getByLabelText("Address of the site in git"), "git@example.org:korenaar/site.git");
    await user.click(screen.getByRole("button", { name: "Bring it in" }));

    expect(await screen.findByText(/Quiqr model found/)).toBeInTheDocument();
    expect(screen.getByText("Ready to edit")).toBeInTheDocument();
  });

  it("shows the key to copy and never a private one", async () => {
    vi.stubGlobal(
      "fetch",
      server({
        [`${base}/key`]: {
          publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIexample chatcms",
          advice: "Add this key to the repository's deploy keys on your git host.",
        },
      }),
    );
    renderPage();

    expect(await screen.findByText(/ssh-ed25519 AAAAC3/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Copy/ })).toBeInTheDocument();
    expect(document.body.textContent).not.toContain("PRIVATE KEY");
  });

  it("says what to do when the key has not been added yet", async () => {
    const fetchMock = vi.fn().mockImplementation((path: string, options?: { method?: string }) => {
      if (options?.method === "POST" && path === `${base}/connection`) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({
            good: false,
            message: "The host answered but would not accept this key. Add the key below.",
          }),
        });
      }
      return Promise.resolve({
        ok: true,
        status: 200,
        json: async () => ({ kind: "none", publicKey: "", advice: "", to: "main-line", form: "source" }),
      });
    });
    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "Test the connection" }));

    expect(await screen.findByRole("status")).toHaveTextContent("would not accept this key");
  });

  it("offers where and what to publish", async () => {
    vi.stubGlobal("fetch", server());
    renderPage();

    const where = await screen.findByRole("radiogroup", { name: "Where to publish" });
    expect(within(where).getByRole("radio", { name: /The site's main line/ })).toBeInTheDocument();
    expect(within(where).getByRole("radio", { name: /A line of its own/ })).toBeInTheDocument();

    const what = screen.getByRole("radiogroup", { name: "What to publish" });
    expect(within(what).getByRole("radio", { name: /The site's files/ })).toBeInTheDocument();
    expect(within(what).getByRole("radio", { name: /The built site/ })).toBeInTheDocument();
  });

  it("says where steps of your own will live", async () => {
    vi.stubGlobal("fetch", server());
    renderPage();

    expect(await screen.findByText(/quiqr\/workflows/)).toBeInTheDocument();
  });

  it("shows what the server said when bringing a site in fails", async () => {
    const fetchMock = vi.fn().mockImplementation((path: string, options?: { method?: string }) => {
      if (options?.method === "POST" && path === `${base}/import`) {
        return Promise.resolve({
          ok: false,
          status: 409,
          json: async () => ({ error: { code: "refused", message: "That address could not be reached." } }),
        });
      }
      return Promise.resolve({
        ok: true,
        status: 200,
        json: async () => ({ kind: "none", publicKey: "", advice: "", to: "main-line", form: "source" }),
      });
    });
    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    renderPage();

    await screen.findByRole("radiogroup", { name: "Where the site comes from" });
    await user.type(screen.getByLabelText("Address of the site in git"), "nowhere");
    await user.click(screen.getByRole("button", { name: "Bring it in" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("That address could not be reached.");
  });
});
