import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { PeoplePage } from "../PeoplePage";
import { ProjectsPage } from "../ProjectsPage";

type Answer = { ok?: boolean; status?: number; json: () => Promise<unknown> };

function answers(byPath: Record<string, Answer>) {
  return vi.fn().mockImplementation((path: string) => {
    const answer = byPath[path];
    if (!answer) {
      return Promise.resolve({ ok: false, status: 404, json: async () => ({ error: { code: "not_found", message: "There is nothing here." } }) });
    }
    return Promise.resolve({ ok: answer.ok ?? true, status: answer.status ?? 200, json: answer.json });
  });
}

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("the projects screen", () => {
  function setUp(projects: unknown[]) {
    vi.stubGlobal(
      "fetch",
      answers({
        "/api/organisations": { json: async () => ({ organisations: [{ slug: "technative", name: "TechNative" }] }) },
        "/api/organisations/technative/projects": { json: async () => ({ projects }) },
      }),
    );

    return render(
      <MemoryRouter initialEntries={["/projects"]}>
        <ProjectsPage />
      </MemoryRouter>,
    );
  }

  it("lists the projects it is given", async () => {
    setUp([
      { slug: "korenaar", name: "Korenaar Bakery", domain: "korenaar.example", yourRole: "owner", youMay: [], unpublished: 4 },
      { slug: "studio-noord", name: "Studio Noord", domain: "studionoord.example", yourRole: "editor", youMay: [], unpublished: 0 },
    ]);

    expect(await screen.findByRole("link", { name: /Korenaar Bakery/ })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Studio Noord/ })).toBeInTheDocument();
    expect(screen.getByText("4 unpublished changes")).toBeInTheDocument();
    expect(screen.getByText("Published")).toBeInTheDocument();
  });

  it("offers a way to start a new one", async () => {
    setUp([]);

    expect(await screen.findByRole("button", { name: /New project/ })).toBeInTheDocument();
    expect(screen.getByText(/Import a site or start from a template/)).toBeInTheDocument();
  });

  it("narrows the list as you search", async () => {
    const user = userEvent.setup();
    setUp([
      { slug: "korenaar", name: "Korenaar Bakery", domain: "korenaar.example", yourRole: "owner", youMay: [], unpublished: 0 },
      { slug: "studio-noord", name: "Studio Noord", domain: "studionoord.example", yourRole: "owner", youMay: [], unpublished: 0 },
    ]);

    await screen.findByRole("link", { name: /Korenaar Bakery/ });
    await user.type(screen.getByRole("textbox", { name: "Search projects" }), "noord");

    expect(screen.queryByRole("link", { name: /Korenaar Bakery/ })).toBeNull();
    expect(screen.getByRole("link", { name: /Studio Noord/ })).toBeInTheDocument();
  });

  it("says so when the list cannot be loaded", async () => {
    vi.stubGlobal(
      "fetch",
      answers({
        "/api/organisations": { json: async () => ({ organisations: [{ slug: "technative", name: "TechNative" }] }) },
        "/api/organisations/technative/projects": {
          ok: false,
          status: 403,
          json: async () => ({ error: { code: "forbidden", message: "You are not allowed to do that." } }),
        },
      }),
    );

    render(
      <MemoryRouter initialEntries={["/projects"]}>
        <ProjectsPage />
      </MemoryRouter>,
    );

    expect(await screen.findByRole("alert")).toHaveTextContent("You are not allowed to do that.");
  });
});

describe("the people and roles screen", () => {
  const people = [
    { personId: "1", email: "joris@example.org", displayName: "Joris Bakker", role: "owner", stillToAccept: false },
    { personId: "2", email: "sam@example.org", displayName: "", role: "viewer", stillToAccept: true },
  ];

  function setUp(fetchMock = answers({
    "/api/organisations/technative/projects/korenaar/people": { json: async () => ({ people }) },
  })) {
    vi.stubGlobal("fetch", fetchMock);

    return render(
      <MemoryRouter initialEntries={["/p/korenaar/people"]}>
        <Routes>
          <Route path="/p/:project/people" element={<PeoplePage organisation="technative" />} />
        </Routes>
      </MemoryRouter>,
    );
  }

  it("lists the people and what they may do", async () => {
    setUp();

    expect(await screen.findByText("Joris Bakker")).toBeInTheDocument();
    // Someone with no name of their own is shown by their address, which then
    // appears both as the name and beneath it.
    expect(screen.getAllByText("sam@example.org").length).toBeGreaterThan(0);
    expect(screen.getByText("Still to accept")).toBeInTheDocument();

    const explanations = screen.getByRole("complementary");
    for (const role of ["Owner", "Implementer", "Editor", "Viewer"]) {
      expect(within(explanations).getByText(role)).toBeInTheDocument();
    }
  });

  it("invites someone with a role", async () => {
    const fetchMock = answers({
      "/api/organisations/technative/projects/korenaar/people": { json: async () => ({ people }) },
    });
    const user = userEvent.setup();
    setUp(fetchMock);

    await screen.findByText("Joris Bakker");

    await user.type(screen.getByLabelText("Invite someone"), "new@example.org");
    await user.selectOptions(screen.getByLabelText("Role for the person you invite"), "viewer");
    await user.click(screen.getByRole("button", { name: "Send invite" }));

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/organisations/technative/projects/korenaar/people",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ email: "new@example.org", role: "viewer" }),
      }),
    );
  });

  it("shows what the server said when an invitation is refused", async () => {
    const fetchMock = vi.fn().mockImplementation((_path: string, options?: { method?: string }) => {
      if (options?.method === "POST") {
        return Promise.resolve({
          ok: false,
          status: 409,
          json: async () => ({ error: { code: "refused", message: "That is not a role." } }),
        });
      }
      return Promise.resolve({ ok: true, status: 200, json: async () => ({ people }) });
    });

    const user = userEvent.setup();
    setUp(fetchMock);
    await screen.findByText("Joris Bakker");

    await user.type(screen.getByLabelText("Invite someone"), "new@example.org");
    await user.click(screen.getByRole("button", { name: "Send invite" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("That is not a role.");
  });

  it("changes a role through a select rather than a menu", async () => {
    const fetchMock = answers({
      "/api/organisations/technative/projects/korenaar/people": { json: async () => ({ people }) },
    });
    const user = userEvent.setup();
    setUp(fetchMock);

    await screen.findByText("Joris Bakker");
    await user.selectOptions(screen.getByLabelText("Role for Joris Bakker"), "implementer");

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/organisations/technative/projects/korenaar/people/1",
      expect.objectContaining({ method: "PUT" }),
    );
  });
});
