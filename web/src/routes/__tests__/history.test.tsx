import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { HistoryPage } from "../HistoryPage";

const base = "/api/organisations/technative/projects/korenaar";

const aTimeline = {
  unpublished: 2,
  groups: [
    {
      period: "Today",
      entries: [
        {
          version: "ccc",
          title: "Spring prices",
          by: "Anna",
          at: "2026-03-20T11:00:00Z",
          changes: 3,
          live: true,
        },
      ],
    },
    {
      period: "Earlier this week",
      entries: [
        {
          version: "aaa",
          title: "The first update",
          by: "Bram",
          at: "2026-03-16T09:00:00Z",
          changes: 1,
          live: false,
        },
      ],
    },
  ],
};

function server(answers: Record<string, unknown>, sent?: (path: string, body: unknown) => void) {
  return vi.fn().mockImplementation((path: string, options?: RequestInit) => {
    if (options?.body) {
      sent?.(path, JSON.parse(String(options.body)));
    }
    const key = options?.method && options.method !== "GET" ? `${options.method} ${path}` : path;
    return Promise.resolve({
      ok: true,
      status: 200,
      json: async () => answers[key] ?? answers[path] ?? {},
    });
  });
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/p/korenaar/history"]}>
      <Routes>
        <Route path="/p/:project/history" element={<HistoryPage organisation="technative" />} />
      </Routes>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("the history screen", () => {
  it("groups versions by period and marks the one that is live", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/versions`]: aTimeline }));
    renderPage();

    expect(await screen.findByText("Today")).toBeInTheDocument();
    expect(screen.getByText("Earlier this week")).toBeInTheDocument();
    expect(screen.getByText("Spring prices")).toBeInTheDocument();
    expect(screen.getByText("Live now")).toBeInTheDocument();
  });

  it("says what has not been published yet", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/versions`]: aTimeline }));
    renderPage();

    expect(await screen.findByText(/Not published yet/)).toBeInTheDocument();
    expect(screen.getByText(/2 changes/)).toBeInTheDocument();
  });

  it("shows what a version contains when asked", async () => {
    vi.stubGlobal(
      "fetch",
      server({
        [`${base}/versions`]: aTimeline,
        [`${base}/versions/aaa`]: {
          count: 1,
          changes: [
            { kind: "changed-page", section: "Pages", page: "Home page", fields: [{ name: "Introduction" }] },
          ],
        },
      }),
    );
    renderPage();

    const buttons = await screen.findAllByRole("button", { name: "See what is in it" });
    await userEvent.click(buttons[1]!);

    expect(await screen.findByText("Home page")).toBeInTheDocument();
    expect(screen.getByText(/Introduction/)).toBeInTheDocument();
  });

  it("offers no way back to the version that is already live", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/versions`]: aTimeline }));
    renderPage();

    await screen.findByText("Spring prices");
    const ways = screen.getAllByRole("button", { name: /Go back to this version/ });
    expect(ways).toHaveLength(1);
    expect(ways[0]).toHaveAccessibleName("Go back to this version…");
  });

  it("says what will happen before going back, and warns about unpublished work", async () => {
    vi.stubGlobal(
      "fetch",
      server({
        [`${base}/versions`]: aTimeline,
        [`${base}/versions/aaa/go-back`]: {
          title: "The first update",
          when: "2026-03-16T09:00:00Z",
          setAside: 2,
          reassurance:
            "2 changes you have not published yet will be set aside. Nothing is lost. Everything stays in your history, and you can undo this afterwards.",
        },
      }),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: /Go back to this version/ }));

    const dialog = await screen.findByRole("dialog");
    expect(dialog).toHaveTextContent("The first update");
    expect(dialog).toHaveTextContent("Nothing is lost");
    expect(dialog).toHaveTextContent("set aside");
  });

  it("publishes right away unless that is turned off", async () => {
    const sent: Array<[string, unknown]> = [];
    vi.stubGlobal(
      "fetch",
      server(
        {
          [`${base}/versions`]: aTimeline,
          [`${base}/versions/aaa/go-back`]: {
            title: "The first update",
            setAside: 0,
            reassurance: "Nothing is lost. Everything stays in your history, and you can undo this afterwards.",
          },
          [`POST ${base}/versions/aaa/go-back`]: {
            title: "The first update",
            published: { id: "a-publish", state: "succeeded", changeCount: 1, startedAt: "2026-03-20T12:00:00Z" },
          },
        },
        (path, body) => sent.push([path, body]),
      ),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: /Go back to this version/ }));
    await screen.findByRole("dialog");

    expect(screen.getByLabelText("Publish right away")).toBeChecked();
    await userEvent.click(screen.getByRole("button", { name: "Yes, go back" }));

    expect(sent).toContainEqual([`${base}/versions/aaa/go-back`, { publishRightAway: true }]);
    expect(await screen.findByText(/It is live now/)).toBeInTheDocument();
  });

  it("can go back without publishing", async () => {
    const sent: Array<[string, unknown]> = [];
    vi.stubGlobal(
      "fetch",
      server(
        {
          [`${base}/versions`]: aTimeline,
          [`${base}/versions/aaa/go-back`]: {
            title: "The first update",
            setAside: 0,
            reassurance: "Nothing is lost.",
          },
          [`POST ${base}/versions/aaa/go-back`]: { title: "The first update" },
        },
        (path, body) => sent.push([path, body]),
      ),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: /Go back to this version/ }));
    await screen.findByRole("dialog");
    await userEvent.click(screen.getByLabelText("Publish right away"));
    await userEvent.click(screen.getByRole("button", { name: "Yes, go back" }));

    expect(sent).toContainEqual([`${base}/versions/aaa/go-back`, { publishRightAway: false }]);
    expect(await screen.findByText(/waiting to be published/)).toBeInTheDocument();
  });

  it("can be called off", async () => {
    vi.stubGlobal(
      "fetch",
      server({
        [`${base}/versions`]: aTimeline,
        [`${base}/versions/aaa/go-back`]: { title: "The first update", setAside: 0, reassurance: "Nothing is lost." },
      }),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: /Go back to this version/ }));
    await screen.findByRole("dialog");
    await userEvent.click(screen.getByRole("button", { name: "Keep things as they are" }));

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("uses no git word anywhere an editor reads", async () => {
    vi.stubGlobal(
      "fetch",
      server({
        [`${base}/versions`]: aTimeline,
        [`${base}/versions/aaa/go-back`]: {
          title: "The first update",
          setAside: 2,
          reassurance: "2 changes you have not published yet will be set aside. Nothing is lost.",
        },
      }),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: /Go back to this version/ }));
    await screen.findByRole("dialog");

    const words = (document.body.textContent ?? "").toLowerCase();
    for (const word of [
      "commit", "branch", "push", "pull", "merge", "repository", "repo",
      "revert", "reset", "head", "sha", "conflict", "remote", "git",
    ]) {
      expect(words).not.toContain(word);
    }
  });

  it("says when a site has no history at all", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/versions`]: { unpublished: 0, groups: [] } }));
    renderPage();

    expect(await screen.findByText(/Nothing has been saved/)).toBeInTheDocument();
  });
});
