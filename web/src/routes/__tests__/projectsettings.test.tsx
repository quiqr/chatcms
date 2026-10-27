import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, describe, expect, it, vi } from "vitest";

import { Inside } from "./here";
import { ProjectSettingsPage } from "../ProjectSettingsPage";

const base = "/api/organisations/technative/projects/korenaar";

function aProject(youMay: string[], yourRole = "owner") {
  return {
    slug: "korenaar",
    name: "Korenaar Bakery",
    domain: "korenaar.example",
    yourRole,
    youMay,
    unpublished: 2,
  };
}

function server(project: unknown, watch?: (path: string, method: string, body: unknown) => void) {
  return vi.fn((path: string, options?: RequestInit) => {
    watch?.(path, options?.method ?? "GET", options?.body ? JSON.parse(String(options.body)) : null);

    if (path === base && (options?.method ?? "GET") === "GET") {
      return Promise.resolve({ ok: true, status: 200, json: async () => project });
    }
    if (path === base && options?.method === "PUT") {
      return Promise.resolve({ ok: true, status: 200, json: async () => project });
    }
    if (path === base && options?.method === "DELETE") {
      return Promise.resolve({ ok: true, status: 204, json: async () => null });
    }
    // The panels ask for the key and how it is published.
    return Promise.resolve({
      ok: true,
      status: 200,
      json: async () => ({ publicKey: "", advice: "Make a key to continue.", to: "main-line", form: "source" }),
    });
  });
}

function renderPage() {
  return render(
    <Inside>
      <MemoryRouter initialEntries={["/o/technative/p/korenaar/settings"]}>
        <Routes>
          <Route path="/o/:organisation/p/:project/settings" element={<ProjectSettingsPage />} />
          <Route path="/o/:organisation/projects" element={<p>the projects</p>} />
        </Routes>
      </MemoryRouter>
    </Inside>,
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
  document.cookie = "chatcms_request_token=; max-age=0; path=/";
});

describe("the project settings screen", () => {
  it("shows what the project is called and where it can be visited", async () => {
    vi.stubGlobal("fetch", server(aProject(["change-settings"])));
    renderPage();

    expect(await screen.findByLabelText("Name")).toHaveValue("Korenaar Bakery");
    expect(screen.getByLabelText("Address shown to customers")).toHaveValue("korenaar.example");
  });

  it("says that the address it is reached at does not change", async () => {
    vi.stubGlobal("fetch", server(aProject(["change-settings"])));
    renderPage();

    expect(
      await screen.findByText(/links already sent keep working/),
    ).toBeInTheDocument();
  });

  it("saves a new name", async () => {
    document.cookie = "chatcms_request_token=a-token; path=/";
    const sent: Array<[string, string, unknown]> = [];
    vi.stubGlobal("fetch", server(aProject(["change-settings"]), (p, m, b) => sent.push([p, m, b])));

    const user = userEvent.setup();
    renderPage();

    await user.clear(await screen.findByLabelText("Name"));
    await user.type(screen.getByLabelText("Name"), "Korenaar BV");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => {
      expect(sent).toContainEqual([
        base,
        "PUT",
        { name: "Korenaar BV", domain: "korenaar.example" },
      ]);
    });
  });

  it("refuses somebody who may not change the settings", async () => {
    vi.stubGlobal("fetch", server(aProject(["read-content"], "editor")));
    renderPage();

    expect(await screen.findByRole("alert")).toHaveTextContent("not allowed");
    expect(screen.queryByLabelText("Name")).toBeNull();
  });

  it("keeps the technical panels from somebody who may not see them", async () => {
    vi.stubGlobal("fetch", server(aProject(["change-settings"], "editor")));
    renderPage();

    await screen.findByLabelText("Name");
    expect(screen.queryByText("Where the site lives")).toBeNull();
    expect(screen.queryByText("How it is published")).toBeNull();
  });

  it("says what removing does and what it leaves alone", async () => {
    vi.stubGlobal("fetch", server(aProject(["change-settings", "delete-project"])));

    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "Remove this project..." }));

    expect(screen.getByText(/The site itself is not touched/)).toBeInTheDocument();
    expect(screen.getByText(/nothing is deleted from where the site lives/)).toBeInTheDocument();
  });

  it("does not offer removing to somebody who may not", async () => {
    vi.stubGlobal("fetch", server(aProject(["change-settings"])));
    renderPage();

    await screen.findByLabelText("Name");
    expect(screen.queryByRole("button", { name: /Remove this project/ })).toBeNull();
  });

  it("goes back to the projects once one is removed", async () => {
    document.cookie = "chatcms_request_token=a-token; path=/";
    vi.stubGlobal("fetch", server(aProject(["change-settings", "delete-project"])));

    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "Remove this project..." }));
    await user.click(screen.getByRole("button", { name: "Yes, remove it" }));

    expect(await screen.findByText("the projects")).toBeInTheDocument();
  });
});
