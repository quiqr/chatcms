import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it } from "vitest";

import { App } from "../../App";
import { PLACEHOLDER, demoProjects } from "../../demo/content";
import { pathFor, screens } from "../screens";

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

  it("sends the bare address to the projects", () => {
    at("/");
    expect(screen.getByRole("button", { name: /TechNative/ })).toBeInTheDocument();
  });

  it("renders both frames so the stylesheet can choose", () => {
    at("/p/korenaar/content");

    expect(screen.getByTestId("project-shell")).toBeInTheDocument();
    expect(screen.getByTestId("phone-shell")).toBeInTheDocument();
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
