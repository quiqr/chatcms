import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { afterEach, describe, expect, it, vi } from "vitest";

import { PlatformUsers } from "../PlatformUsers";
import { Inside } from "./here";

function answers(byPath: Record<string, unknown>) {
  return vi.fn((path: string) => {
    const body = byPath[path] ?? byPath[path.split("?")[0]!];
    if (body === undefined) {
      return Promise.resolve({ ok: false, status: 404, json: async () => ({}) });
    }
    return Promise.resolve({ ok: true, status: 200, json: async () => body });
  });
}

const somebody = {
  personId: "p1",
  email: "marieke@korenaar.example",
  displayName: "Marieke de Vries",
  administers: false,
  accountOn: true,
  stillToAccept: false,
  organisations: [{ slug: "bakery", name: "Korenaar Bakery", role: "admin" }],
  projects: [],
};

function withPeople(people: unknown[], rest: Record<string, unknown> = {}) {
  return answers({
    "/api/administration/people": { people, shown: people.length, altogether: people.length },
    "/api/administration/organisations": { organisations: [] },
    ...rest,
  });
}

function show() {
  return render(
    <MemoryRouter>
      <Inside>
        <PlatformUsers />
      </Inside>
    </MemoryRouter>,
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("every account", () => {
  it("says which of the three states each account is in", async () => {
    vi.stubGlobal(
      "fetch",
      withPeople([
        somebody,
        { ...somebody, personId: "p2", email: "invited@example.org", displayName: "", stillToAccept: true },
        { ...somebody, personId: "p3", email: "blocked@example.org", displayName: "", accountOn: false },
      ]),
    );

    show();

    // Only one of the three is a column: blocked. Invited is "never set a
    // password", and active is neither.
    expect(await screen.findByText("Active")).toBeInTheDocument();
    expect(screen.getAllByText("Invited").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Blocked").length).toBeGreaterThan(0);
  });

  it("says where somebody belongs, and when they belong nowhere", async () => {
    vi.stubGlobal(
      "fetch",
      withPeople([somebody, { ...somebody, personId: "p2", email: "nowhere@example.org", organisations: [] }]),
    );

    show();

    expect(await screen.findByText(/Korenaar Bakery \(Organisation admin\)/)).toBeInTheDocument();
    expect(screen.getAllByText("Not in any organisation").length).toBeGreaterThan(0);
  });

  it("offers nothing dangerous against your own account", async () => {
    // Inside() signs the test in as joris@example.org.
    vi.stubGlobal(
      "fetch",
      withPeople([{ ...somebody, personId: "p9", email: "joris@example.org", displayName: "Joris" }]),
    );

    const user = userEvent.setup();
    show();

    await user.click(await screen.findByRole("button", { name: /What can be done with Joris/ }));

    expect(await screen.findByRole("menuitem", { name: "Block sign-in" })).toHaveAttribute(
      "aria-disabled",
      "true",
    );
    expect(screen.getByRole("menuitem", { name: /Delete user/ })).toHaveAttribute(
      "aria-disabled",
      "true",
    );
    // Changing their own name is theirs to do, and is not refused.
    expect(screen.getByRole("menuitem", { name: "Edit name and email" })).not.toHaveAttribute(
      "aria-disabled",
      "true",
    );
  });

  it("asks before deleting, and says what stays", async () => {
    vi.stubGlobal("fetch", withPeople([somebody]));

    const user = userEvent.setup();
    show();

    await user.click(await screen.findByRole("button", { name: /What can be done with Marieke/ }));
    await user.click(screen.getByRole("menuitem", { name: /Delete user/ }));

    expect(await screen.findByText(/What that person wrote stays/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Yes, delete the account" })).toBeInTheDocument();
  });
});
