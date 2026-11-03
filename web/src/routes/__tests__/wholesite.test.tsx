import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { HereContext, type WhoAmI } from "../../lib/whoami";
import type { OrganisationArea, Project } from "../../lib/tenancy";
import { ManageProjects } from "../ManageProjects";

function site(slug: string, name: string, archived = false): Project {
  return {
    slug,
    name,
    domain: `${slug}.example`,
    yourRole: "implementer",
    youMay: [],
    unpublished: 0,
    archived,
  };
}

function area(over: Partial<OrganisationArea> = {}): OrganisationArea {
  return {
    slug: "technative",
    name: "TechNative",
    usage: { members: { held: 2, full: false }, projects: { held: 1, full: false } },
    people: [],
    projects: [site("korenaar", "Korenaar Bakery")],
    ...over,
  };
}

const who: WhoAmI = {
  person: { id: "person-1", email: "joris@example.org", displayName: "Joris Bakker" },
  administers: false,
  organisations: [
    { slug: "technative", name: "TechNative", youAdminister: true },
    { slug: "stichting", name: "Stichting Broodnodig", youAdminister: true },
    { slug: "elders", name: "Elders BV", youAdminister: false },
  ],
};

function show(it: OrganisationArea, again = () => {}) {
  return render(
    <HereContext.Provider value={{ who, organisation: it.slug, again: () => {} }}>
      <MemoryRouter>
        <ManageProjects area={it} again={again} />
      </MemoryRouter>
    </HereContext.Provider>,
  );
}

function answered(body: unknown = {}) {
  const fetched = vi.fn(() =>
    Promise.resolve({ ok: true, status: 200, json: async () => body }),
  );
  vi.stubGlobal("fetch", fetched);
  return fetched;
}

async function openTheMenu(user: ReturnType<typeof userEvent.setup>, name: string) {
  await user.click(screen.getByRole("button", { name: `What can be done with ${name}` }));
}

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("what a whole site can be put through", () => {
  it("offers the four acts instead of saying they are not built", async () => {
    const user = userEvent.setup();
    show(area());

    await openTheMenu(user, "Korenaar Bakery");

    expect(screen.queryByText("Not built yet")).toBeNull();
    expect(screen.getByRole("menuitem", { name: "Duplicate..." })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Move to another organisation..." })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Archive" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Download as zip" })).toHaveAttribute(
      "href",
      "/api/organisations/technative/projects/korenaar/zip",
    );
  });

  it("starts a copy from the original's name, and refuses it until it differs", async () => {
    const user = userEvent.setup();
    const fetched = answered(site("korenaar-2", "Korenaar Bakery 2"));
    show(area());

    await openTheMenu(user, "Korenaar Bakery");
    await user.click(screen.getByRole("menuitem", { name: "Duplicate..." }));

    const name = screen.getByRole("textbox", { name: /Name of the copy/ });
    expect(name).toHaveValue("Korenaar Bakery");
    expect(name).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByRole("button", { name: "Duplicate site" })).toBeDisabled();

    await user.clear(name);
    await user.type(name, "Korenaar Bakery 2");
    await user.click(screen.getByLabelText(/People and their roles/));
    await user.click(screen.getByRole("button", { name: "Duplicate site" }));

    await waitFor(() => expect(fetched).toHaveBeenCalled());
    const [where, how] = fetched.mock.calls[0] as unknown as [string, { body: string }];
    expect(where).toBe("/api/organisations/technative/projects/korenaar/copies");
    expect(JSON.parse(how.body)).toEqual({
      slug: "korenaar-bakery-2",
      name: "Korenaar Bakery 2",
      withPeople: true,
      withPublishing: false,
    });
  });

  it("lets a copy go only where there is room", async () => {
    const user = userEvent.setup();
    show(area({ usage: { members: { held: 2, full: false }, projects: { held: 3, allowed: 3, full: true } } }));

    await openTheMenu(user, "Korenaar Bakery");
    await user.click(screen.getByRole("menuitem", { name: "Duplicate..." }));

    expect(screen.getByRole("button", { name: "Duplicate site" })).toBeDisabled();
  });

  it("moves a site only to an organisation you administer", async () => {
    const user = userEvent.setup();
    const fetched = answered();
    show(area());

    await openTheMenu(user, "Korenaar Bakery");
    await user.click(screen.getByRole("menuitem", { name: "Move to another organisation..." }));

    expect(screen.getByRole("radio", { name: /Stichting Broodnodig/ })).toBeEnabled();
    expect(screen.getByRole("radio", { name: /Elders BV/ })).toBeDisabled();
    expect(screen.queryByRole("radio", { name: /TechNative/ })).toBeNull();

    await user.click(screen.getByRole("button", { name: "Move site" }));

    await waitFor(() => expect(fetched).toHaveBeenCalled());
    const [where, how] = fetched.mock.calls[0] as unknown as [string, { body: string }];
    expect(where).toBe("/api/organisations/technative/projects/korenaar/organisation");
    expect(JSON.parse(how.body)).toEqual({ to: "stichting" });
  });

  it("puts archived sites in their own group and brings one back", async () => {
    const user = userEvent.setup();
    const fetched = answered();
    const again = vi.fn();
    show(
      area({ projects: [site("korenaar", "Korenaar Bakery"), site("oud", "Oud Nieuws", true)] }),
      again,
    );

    const group = screen.getByRole("heading", { name: /Archived · 1/ });
    expect(group).toBeInTheDocument();
    expect(within(group).getByText(/doesn’t count toward the limit/)).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /Oud Nieuws/ })).toBeNull();

    await user.click(screen.getByRole("button", { name: "Bring it back" }));

    await waitFor(() => expect(again).toHaveBeenCalled());
    const [where, how] = fetched.mock.calls[0] as unknown as [string, { method: string; body: string }];
    expect(where).toBe("/api/organisations/technative/projects/oud/archived");
    expect(how.method).toBe("PUT");
    expect(JSON.parse(how.body)).toEqual({ archived: false });
  });

  it("says what went wrong when an act is refused", async () => {
    const user = userEvent.setup();
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 409,
          json: async () => ({ error: { code: "no_room_for_a_site", message: "There is no room for another site." } }),
        }),
      ),
    );
    show(area());

    await openTheMenu(user, "Korenaar Bakery");
    await user.click(screen.getByRole("menuitem", { name: "Archive" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/no room/i);
  });
});
