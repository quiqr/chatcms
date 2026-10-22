import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { App } from "../../App";
import { PLACEHOLDER, demoProjects } from "../../demo/content";
import { pathFor, screens } from "../screens";

beforeEach(() => {
  // Screens that read from the server get an empty answer, so these tests are
  // about addresses rather than about what any screen shows.
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ organisations: [], projects: [], people: [] }) }),
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
  it("gives every milestone screen an address that reaches it", () => {
    for (const screen_ of screens) {
      const path = pathFor(screen_.name);
      const { unmount } = at(path);

      expect(
        document.querySelector(
          "[data-testid='project-shell'], [data-testid='phone-shell'], [data-testid='organisation-shell'], [data-testid='editor-shell'], main",
        ),
        `${screen_.name} at ${path} renders nothing`,
      ).not.toBeNull();

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

  it("renders one frame rather than both, so a screen loads once", () => {
    at("/p/korenaar/content");

    // The test environment reports a wide window, so the project frame is the
    // one chosen. Which frame is chosen for which width is the shells' own
    // test; what matters here is that only one is mounted.
    expect(screen.getByTestId("project-shell")).toBeInTheDocument();
    expect(screen.queryByTestId("phone-shell")).toBeNull();
  });
});

describe("placeholder data", () => {
  it("is in one place and says so", () => {
    expect(PLACEHOLDER).toBe(true);
    expect(demoProjects.length).toBeGreaterThan(0);
  });

  it("refuses to name a screen that does not exist", () => {
    // @ts-expect-error the point of the test is the unknown name
    expect(() => pathFor("not-a-screen")).toThrow();
  });
});
