import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { App } from "../../App";
import { pathFor, screens } from "../screens";

beforeEach(() => {
  // Screens that read from the server get an empty answer, so these tests are
  // about addresses rather than about what any screen shows.
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      // One answer that satisfies every screen, because this test is about
      // addresses rather than about what any screen shows.
      json: async () => ({
        person: { id: "person-1", email: "joris@example.org", displayName: "Joris Bakker" },
        administers: false,
        organisations: [{ slug: "technative", name: "TechNative" }],
        projects: [],
        // Also a project, because the shell reads the real one now.
        slug: "korenaar",
        name: "Korenaar Bakery",
        domain: "korenaar.example",
        yourRole: "owner",
        youMay: ["read-content", "edit-content", "publish", "change-settings"],
        unpublished: 0,
        people: [],
        sections: [],
        images: [],
        versions: [],
        changes: [],
        records: [],
        groups: [],
        fields: [],
        body: "",
        title: "",
        section: "",
        hasBody: false,
      }),
    }),
  );
});

afterEach(() => {
  vi.unstubAllGlobals();
});

function at(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <App />
    </MemoryRouter>,
  );
}

describe("routes", () => {
  it("gives every milestone screen an address that reaches it", async () => {
    for (const screen_ of screens) {
      const path = pathFor(screen_.name, "technative");
      const { unmount } = at(path);

      // The shell asks the server who is looking before it draws anything,
      // so this waits for the screen rather than asserting against the wait.
      await waitFor(
        () => {
          expect(
            document.querySelector(
              "[data-testid='project-shell'], [data-testid='phone-shell'], [data-testid='organisation-shell'], [data-testid='editor-shell'], main",
            ),
            `${screen_.name} at ${path} renders nothing`,
          ).not.toBeNull();
        },
        { timeout: 3000 },
      );

      expect(screen.queryByText("This address does not exist"), `${screen_.name} at ${path}`).toBeNull();
      unmount();
    }
  });

  it("explains an address that names nothing, and offers a way back", () => {
    at("/nowhere-at-all");

    expect(screen.getByRole("heading", { name: "This address does not exist" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Go to your projects" })).toBeInTheDocument();
  });

  it("sends the bare address to the projects", async () => {
    at("/");
    expect(await screen.findByRole("heading", { name: "Projects" })).toBeInTheDocument();
  });

  it("renders one frame rather than both, so a screen loads once", async () => {
    at("/o/technative/p/korenaar/content");

    // The test environment reports a wide window, so the project frame is the
    // one chosen. Which frame is chosen for which width is the shells' own
    // test; what matters here is that only one is mounted.
    expect(await screen.findByTestId("project-shell")).toBeInTheDocument();
    expect(screen.queryByTestId("phone-shell")).toBeNull();
  });
});
