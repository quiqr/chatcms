import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { afterEach, describe, expect, it, vi } from "vitest";

import { AdministrationPage } from "../AdministrationPage";

function answers(byPath: Record<string, unknown>) {
  return vi.fn((path: string) => {
    const body = byPath[path] ?? byPath[path.split("?")[0]!];
    if (body === undefined) {
      return Promise.resolve({ ok: false, status: 404, json: async () => ({}) });
    }
    return Promise.resolve({ ok: true, status: 200, json: async () => body });
  });
}

const empty = {
  "/api/administration/organisations": { organisations: [] },
  "/api/administration/people": { people: [], shown: 0, altogether: 0 },
  "/api/administration/administrators": { administrators: [] },
  "/api/administration/records": { records: [] },
};

function renderPage() {
  return render(
    <MemoryRouter>
      <AdministrationPage />
    </MemoryRouter>,
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
  document.cookie = "chatcms_request_token=; max-age=0; path=/";
});

describe("the administration screen", () => {
  it("lists every organisation, including ones the administrator is not in", async () => {
    vi.stubGlobal(
      "fetch",
      answers({
        ...empty,
        "/api/administration/organisations": {
          organisations: [
            { slug: "bakery", name: "Korenaar Bakery", projects: 2, people: 3 },
            { slug: "studio", name: "Studio Noord", projects: 1, people: 1 },
          ],
        },
      }),
    );

    renderPage();

    expect(await screen.findByText("Korenaar Bakery")).toBeInTheDocument();
    expect(screen.getByText("Studio Noord")).toBeInTheDocument();
    expect(screen.getByText(/2 sites · 3 people/)).toBeInTheDocument();
    expect(screen.getByText(/1 site · 1 person/)).toBeInTheDocument();
  });

  it("says that starting one does not put you in it", async () => {
    vi.stubGlobal("fetch", answers(empty));
    renderPage();

    expect(
      await screen.findByText(/You will not be in it\./),
    ).toBeInTheDocument();
  });

  it("suggests the address from the name", async () => {
    vi.stubGlobal("fetch", answers(empty));
    const user = userEvent.setup();
    renderPage();

    await user.type(await screen.findByLabelText("Name"), "Korenaar Bakery");

    expect(screen.getByText("Its address will be /o/korenaar-bakery")).toBeInTheDocument();
  });

  it("asks who owns it, and sends both", async () => {
    document.cookie = "chatcms_request_token=a-token; path=/";
    const fetchMock = answers(empty);
    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    renderPage();

    await user.type(await screen.findByLabelText("Name"), "Korenaar Bakery");
    await user.type(screen.getByLabelText("Who owns it"), "marieke@korenaar.example");
    await user.click(screen.getByRole("button", { name: "Start it" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/administration/organisations",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            slug: "korenaar-bakery",
            name: "Korenaar Bakery",
            ownerEmail: "marieke@korenaar.example",
          }),
        }),
      );
    });
  });

  it("marks which administrator is you", async () => {
    vi.stubGlobal(
      "fetch",
      answers({
        ...empty,
        "/api/administration/administrators": {
          administrators: [
            { email: "pim@example.org", displayName: "Pim Snel", yourself: true },
            { email: "other@example.org", displayName: "", yourself: false },
          ],
        },
      }),
    );

    renderPage();

    expect(await screen.findByText("Pim Snel")).toBeInTheDocument();
    expect(screen.getByText("You")).toBeInTheDocument();
    expect(screen.getByText("other@example.org")).toBeInTheDocument();
  });

  it("shows what has been done, and says so when nothing has", async () => {
    vi.stubGlobal("fetch", answers(empty));
    renderPage();

    expect(await screen.findByText("Nothing has been done yet.")).toBeInTheDocument();
  });

  it("reads the record in plain language", async () => {
    vi.stubGlobal(
      "fetch",
      answers({
        ...empty,
        "/api/administration/records": {
          records: [
            {
              who: "Pim Snel",
              did: "made an organisation",
              to: "Korenaar Bakery",
              detail: { ownedBy: "marieke@korenaar.example" },
              at: "2026-10-26T09:00:00Z",
            },
          ],
        },
      }),
    );

    renderPage();

    expect(await screen.findByText(/Pim Snel made an organisation/)).toBeInTheDocument();
    expect(screen.getByText(/owned by marieke@korenaar.example/)).toBeInTheDocument();
  });

  it("says what the server said when starting one is refused", async () => {
    document.cookie = "chatcms_request_token=a-token; path=/";
    vi.stubGlobal(
      "fetch",
      vi.fn((path: string, options?: { method?: string }) => {
        // Only the attempt to start one is refused; the listing beside it
        // still answers, which is what the screen looks like in practice.
        if (path === "/api/administration/organisations" && options?.method === "POST") {
          return Promise.resolve({
            ok: false,
            status: 409,
            json: async () => ({
              error: { code: "refused", message: "That address is already taken." },
            }),
          });
        }
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({
            organisations: [],
            administrators: [],
            records: [],
            people: [],
            shown: 0,
            altogether: 0,
          }),
        });
      }),
    );

    const user = userEvent.setup();
    renderPage();

    await user.type(await screen.findByLabelText("Name"), "Korenaar Bakery");
    await user.type(screen.getByLabelText("Who owns it"), "marieke@korenaar.example");
    await user.click(screen.getByRole("button", { name: "Start it" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("That address is already taken.");
  });

  it("lists people with where they belong, and what it is not showing", async () => {
    vi.stubGlobal(
      "fetch",
      answers({
        ...empty,
        "/api/administration/people": {
          people: [
            {
              personId: "p1",
              email: "marieke@korenaar.example",
              displayName: "Marieke de Vries",
              administers: false,
              accountOn: true,
              stillToAccept: false,
              organisations: [{ slug: "bakery", name: "Korenaar Bakery", role: "member" }],
              projects: [
                { slug: "korenaar", name: "The site", organisation: "bakery", role: "editor" },
              ],
            },
          ],
          shown: 1,
          altogether: 12,
        },
      }),
    );

    renderPage();

    expect(await screen.findByText("Marieke de Vries")).toBeInTheDocument();
    expect(screen.getByText(/Korenaar Bakery \(member\)/)).toBeInTheDocument();
    expect(screen.getByText(/The site \(editor\)/)).toBeInTheDocument();
    expect(screen.getByText(/Showing 1 of 12/)).toBeInTheDocument();
  });

  it("says a person was sent a way in, rather than a password", async () => {
    document.cookie = "chatcms_request_token=a-token; path=/";
    vi.stubGlobal(
      "fetch",
      vi.fn((path: string, options?: { method?: string }) => {
        if (path === "/api/administration/people" && options?.method === "POST") {
          return Promise.resolve({
            ok: true,
            status: 201,
            json: async () => ({ email: "sam@example.org" }),
          });
        }
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({
            organisations: [],
            administrators: [],
            records: [],
            people: [],
            shown: 0,
            altogether: 0,
          }),
        });
      }),
    );

    const user = userEvent.setup();
    renderPage();

    await user.type(await screen.findByLabelText("Email"), "sam@example.org");
    await user.click(screen.getByRole("button", { name: "Make them" }));

    expect(await screen.findByRole("status")).toHaveTextContent("sent a way in");
  });
});
