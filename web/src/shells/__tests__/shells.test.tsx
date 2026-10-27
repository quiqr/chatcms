import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { MemoryRouter } from "react-router";
import { describe, expect, it } from "vitest";

import type { Project } from "../../lib/tenancy";
import { EditorShell } from "../EditorShell";
import { OrganisationShell } from "../OrganisationShell";
import { PhoneShell } from "../PhoneShell";
import { ProjectShell } from "../ProjectShell";
import { Inside } from "../../routes/__tests__/here";
import { canSeeSetup, canSeeTechnicalDetail } from "../roles";
import { phoneWidth, useIsPhone } from "../useIsPhone";

// The shells are given a project the way the server describes one.
const theViewer = { name: "Joris Bakker", role: "owner" };

const project: Project = {
  slug: "korenaar",
  name: "Korenaar Bakery",
  domain: "korenaar.example",
  yourRole: "owner",
  youMay: ["read-content", "edit-content", "publish", "change-settings"],
  unpublished: 4,
};

function at(path: string, element: ReactNode) {
  return render(
    <Inside>
      <MemoryRouter initialEntries={[path]}>{element}</MemoryRouter>
    </Inside>,
  );
}

describe("ProjectShell", () => {
  it("shows the site navigation", () => {
    at("/o/technative/p/korenaar/content", <ProjectShell project={project} viewer={theViewer} title="Content">body</ProjectShell>);

    const navigation = screen.getByRole("complementary", { name: "Project navigation" });
    for (const label of ["Content", "Media", "Publish", "History"]) {
      expect(within(navigation).getByRole("link", { name: new RegExp(label) })).toBeInTheDocument();
    }
  });

  it("shows how much is waiting to be published", () => {
    at("/o/technative/p/korenaar/content", <ProjectShell project={project} viewer={theViewer} title="Content">body</ProjectShell>);

    const publish = screen.getAllByRole("link", { name: /Publish/ });
    expect(publish.some((link) => link.textContent?.includes(String(project.unpublished)))).toBe(true);
  });

  it("leaves setup out for an editor rather than showing it greyed out", () => {
    at(
      "/o/technative/p/korenaar/content",
      <ProjectShell project={project} viewer={{ name: "Marieke", role: "editor" }} title="Content">
        body
      </ProjectShell>,
    );

    expect(screen.queryByRole("link", { name: /People & roles/ })).toBeNull();
    expect(screen.queryByRole("link", { name: /Project settings/ })).toBeNull();
    expect(screen.queryByText(/Setup/)).toBeNull();
  });

  it("gives the content the room and lets the navigation yield", () => {
    const { container } = at(
      "/o/technative/p/korenaar/content",
      <ProjectShell project={project} viewer={theViewer} title="Content">body</ProjectShell>,
    );

    const shell = container.querySelector("[data-testid='project-shell']")!;
    expect(shell.className).toContain("flex-wrap");

    const content = shell.querySelector("[class*='grow-[999]']");
    expect(content).not.toBeNull();
    expect(content!.className).toContain("min-w-0");
  });

  it("is hidden at phone widths", () => {
    const { container } = at(
      "/o/technative/p/korenaar/content",
      <ProjectShell project={project} viewer={theViewer} title="Content">body</ProjectShell>,
    );

    const shell = container.querySelector("[data-testid='project-shell']")!;
    expect(shell.className).toContain("hidden");
    expect(shell.className).toContain("sm:flex");
  });
});

describe("PhoneShell", () => {
  it("has a tab bar for the site sections", () => {
    at("/o/technative/p/korenaar/content", <PhoneShell project={project} viewer={theViewer}>body</PhoneShell>);

    const tabs = screen.getByRole("navigation", { name: "Site sections" });
    for (const label of ["Content", "Media", "Publish", "History"]) {
      expect(within(tabs).getByRole("link", { name: new RegExp(label) })).toBeInTheDocument();
    }
  });

  it("keeps its bottom bar clear of the device", () => {
    at("/o/technative/p/korenaar/content", <PhoneShell project={project} viewer={theViewer}>body</PhoneShell>);

    expect(screen.getByTestId("phone-tab-bar").className).toContain("pb-safe");
  });

  it("gives way to the editing actions while a page is being edited", () => {
    at(
      "/o/technative/p/korenaar/content/about",
      <PhoneShell project={project} viewer={theViewer} bottomBar={<button type="button">Publish</button>}>
        body
      </PhoneShell>,
    );

    expect(screen.queryByTestId("phone-tab-bar")).toBeNull();
    expect(screen.getByTestId("phone-action-bar").className).toContain("pb-safe-2");
  });

  it("is hidden above phone widths", () => {
    const { container } = at("/o/technative/p/korenaar/content", <PhoneShell project={project} viewer={theViewer}>body</PhoneShell>);

    expect(container.querySelector("[data-testid='phone-shell']")!.className).toContain("sm:hidden");
  });
});

describe("OrganisationShell", () => {
  it("uses a header rather than the project navigation", () => {
    at(
      "/projects",
      <OrganisationShell organisation={{ name: "TechNative" }} viewer={theViewer}>
        body
      </OrganisationShell>,
    );

    expect(screen.getByRole("button", { name: /TechNative/ })).toBeInTheDocument();
    expect(screen.queryByRole("complementary", { name: "Project navigation" })).toBeNull();
  });
});

describe("EditorShell", () => {
  function editor(layout: "form" | "split" | "preview" = "split") {
    return at(
      "/o/technative/p/korenaar/content/about",
      <EditorShell
        project={project}
        viewer={theViewer}
        breadcrumb="News / About us"
        layout={layout}
        onLayoutChange={() => {}}
        form={<p>Fields</p>}
        preview={<p>Preview</p>}
      />,
    );
  }

  it("has a slot for the form, the preview, and the one that comes later", () => {
    editor();

    expect(screen.getByTestId("editor-slot-form")).toBeInTheDocument();
    expect(screen.getByTestId("editor-slot-preview")).toBeInTheDocument();

    const assistant = screen.getByTestId("editor-slot-assistant");
    expect(assistant).toBeInTheDocument();
    expect(assistant).not.toBeVisible();
  });

  it("uses a narrow rail rather than the full navigation", () => {
    editor();

    const rail = screen.getByTestId("editor-rail");
    expect(rail.className).toContain("w-16");
    expect(within(rail).getByRole("link", { name: "Content" })).toBeInTheDocument();
    expect(screen.queryByRole("complementary", { name: "Project navigation" })).toBeNull();
  });

  it("shows one column when asked for one", () => {
    const form = editor("form");
    expect(screen.queryByTestId("editor-slot-preview")).toBeNull();
    form.unmount();

    editor("preview");
    expect(screen.queryByTestId("editor-slot-form")).toBeNull();
  });

  it("keeps the assistant reachable but says it is not ready", () => {
    editor();
    expect(screen.getByRole("button", { name: /Assistant/ })).toBeDisabled();
  });
});

describe("roles", () => {
  it("decides who sees setup and technical detail", () => {
    const expected: Record<string, boolean> = {
      owner: true,
      implementer: true,
      editor: false,
      viewer: false,
    };

    for (const [role, allowed] of Object.entries(expected)) {
      expect(canSeeSetup(role)).toBe(allowed);
      expect(canSeeTechnicalDetail(role)).toBe(allowed);
    }
  });
});

describe("useIsPhone", () => {
  it("reads the same width the stylesheet uses", () => {
    expect(phoneWidth).toBe(640);
  });

  it("says no when the window is wide", () => {
    const { result } = renderHook(() => useIsPhone());
    expect(result).toBe(false);
  });

  it("says yes when the window is narrow", () => {
    const original = window.matchMedia;
    window.matchMedia = ((query: string) =>
      ({
        matches: true,
        media: query,
        addEventListener: () => {},
        removeEventListener: () => {},
      }) as unknown as MediaQueryList) as typeof window.matchMedia;

    const { result } = renderHook(() => useIsPhone());
    expect(result).toBe(true);

    window.matchMedia = original;
  });
});

function renderHook<T>(use: () => T): { result: T } {
  const box = { result: undefined as unknown as T };

  function Probe() {
    box.result = use();
    return null;
  }

  render(<Probe />);
  return box;
}

describe("the phone frame and sideways scrolling", () => {
  it("clips sideways rather than letting the page move", () => {
    at("/o/technative/p/korenaar/content", <PhoneShell project={project} viewer={theViewer}>body</PhoneShell>);

    const shell = screen.getByTestId("phone-shell");
    expect(shell.className).toContain("overflow-x-clip");

    const main = shell.querySelector("main")!;
    expect(main.className).toContain("overflow-x-clip");
  });
});

describe("the way out", () => {
  // This defect was one shell having the link and three having a picture
  // that did nothing, so the test walks all four rather than the one that
  // happened to be wrong.
  const everyShell: [string, ReactNode][] = [
    ["ProjectShell", <ProjectShell project={project} viewer={theViewer} title="Content">{null}</ProjectShell>],
    ["PhoneShell", <PhoneShell project={project} viewer={theViewer}>{null}</PhoneShell>],
    [
      "EditorShell",
      <EditorShell
        project={project}
        viewer={theViewer}
        breadcrumb="News"
        layout="split"
        onLayoutChange={() => {}}
        form={null}
        preview={null}
      />,
    ],
    [
      "OrganisationShell",
      <OrganisationShell organisation={{ name: "TechNative" }} viewer={theViewer}>{null}</OrganisationShell>,
    ],
  ];

  for (const [name, shell] of everyShell) {
    it(`${name} offers the way to yourself and the way out`, async () => {
      const { unmount } = at("/o/technative/p/korenaar/content", shell);

      const person = screen.getByTestId("person-menu");
      expect(person, `${name} has no person menu`).toBeInTheDocument();

      await userEvent.click(person);

      expect(await screen.findByText("Your account"), `${name} has no way to your account`).toBeInTheDocument();
      expect(screen.getByText("Sign out"), `${name} has no way out`).toBeInTheDocument();
      expect(screen.getByText("Sign out everywhere")).toBeInTheDocument();

      unmount();
    });
  }
});
