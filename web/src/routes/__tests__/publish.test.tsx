import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { PublishPage } from "../PublishPage";

const base = "/api/organisations/technative/projects/korenaar";

const aReview = {
  count: 2,
  lastPublished: "2026-03-01T09:00:00Z",
  changes: [
    {
      kind: "changed-page",
      section: "Pages",
      page: "Home page",
      by: "Anna",
      at: "2026-03-04T11:00:00Z",
      fields: [
        {
          name: "Introduction",
          words: [
            { part: "kept", text: "We sell " },
            { part: "gone", text: "apples " },
            { part: "added", text: "plums " },
            { part: "kept", text: "and pears." },
          ],
        },
      ],
    },
    { kind: "new-image", section: "Images", page: "Logo", by: "Bram" },
  ],
};

function server(answers: Record<string, unknown>, sent?: (body: unknown) => void) {
  return vi.fn().mockImplementation((path: string, options?: RequestInit) => {
    if (options?.body) {
      sent?.(JSON.parse(String(options.body)));
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
    <MemoryRouter initialEntries={["/p/korenaar/publish"]}>
      <Routes>
        <Route path="/p/:project/publish" element={<PublishPage organisation="technative" />} />
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

// gitWords are the words docs/design/DESIGN.md section 7 keeps out of
// anything an editor reads.
const gitWords = [
  "commit", "branch", "push", "pull", "merge", "repository", "repo",
  "deploy", "HEAD", "SHA", "conflict", "remote", "git",
];

function readableText() {
  const technical = document.querySelector("[class*='chatcms-audience'], pre");
  const words = document.body.textContent ?? "";
  return technical ? words.replace(technical.textContent ?? "", "") : words;
}

describe("the publish screen", () => {
  it("says how many changes are waiting and when the site last went out", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/review`]: aReview }));
    renderPage();

    expect(await screen.findByText("Ready to publish 2 changes")).toBeInTheDocument();
    expect(screen.getByText(/Last published/)).toBeInTheDocument();
  });

  it("lists each change with a badge, the page and who made it", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/review`]: aReview }));
    renderPage();

    expect(await screen.findByText("Home page")).toBeInTheDocument();
    expect(screen.getByText("Changed")).toBeInTheDocument();
    expect(screen.getByText("New image")).toBeInTheDocument();
    expect(screen.getByText(/Anna/)).toBeInTheDocument();
  });

  it("shows the words that changed when the change is expanded", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/review`]: aReview }));
    renderPage();

    const expand = await screen.findByRole("button", { name: "Show changes" });
    expect(screen.queryByText("apples")).not.toBeInTheDocument();

    await userEvent.click(expand);

    const removed = screen.getByText("apples");
    const added = screen.getByText("plums");
    expect(removed.tagName).toBe("DEL");
    expect(added.tagName).toBe("INS");
    expect(screen.getAllByText("Introduction").length).toBeGreaterThan(0);
  });

  it("sends the note with the publish", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal(
      "fetch",
      server(
        {
          [`${base}/review`]: aReview,
          [`POST ${base}/publishes`]: {
            id: "a-publish", state: "succeeded", changeCount: 2,
            startedAt: "2026-03-04T12:00:00Z",
            steps: [{ title: "Putting it live", state: "succeeded" }],
          },
        },
        (body) => sent.push(body),
      ),
    );
    renderPage();

    await userEvent.type(await screen.findByLabelText("What's in this update?"), "Spring prices");
    await userEvent.click(screen.getByRole("button", { name: "Publish 2 changes" }));

    expect(sent).toContainEqual({ note: "Spring prices" });
    expect(await screen.findByText("Your site is up to date")).toBeInTheDocument();
  });

  it("says nothing is waiting when nothing is", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/review`]: { count: 0, changes: [] } }));
    renderPage();

    expect(await screen.findByText("Nothing is waiting")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Publish/ })).toBeDisabled();
  });

  it("tells an editor what happened and offers to try again", async () => {
    vi.stubGlobal(
      "fetch",
      server({
        [`${base}/review`]: aReview,
        [`POST ${base}/publishes`]: {
          id: "a-publish", state: "failed", changeCount: 2,
          startedAt: "2026-03-04T12:00:00Z",
          trouble:
            "We couldn't publish this time. Nothing changed on your live site, and your edits are safe. A check on this site found something that needs fixing first.",
          steps: [{ title: "Checking pages and links", state: "failed" }],
        },
      }),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: "Publish 2 changes" }));

    expect(await screen.findByText(/Nothing changed on your live site/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Try again" })).toBeInTheDocument();
  });

  it("keeps the raw output behind a disclosure and out of what an editor reads", async () => {
    vi.stubGlobal(
      "fetch",
      server({
        [`${base}/review`]: aReview,
        [`POST ${base}/publishes`]: {
          id: "a-publish", state: "failed", changeCount: 2,
          startedAt: "2026-03-04T12:00:00Z",
          trouble: "We couldn't publish this time. Nothing changed on your live site, and your edits are safe.",
          detail: "fatal: not a git repository",
          steps: [{ title: "Building the site", state: "failed" }],
        },
      }),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: "Publish 2 changes" }));
    await screen.findByText(/Nothing changed on your live site/);

    expect(screen.queryByText(/not a git repository/)).not.toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: /Technical details/ }));
    expect(screen.getByText(/not a git repository/)).toBeInTheDocument();
  });

  it("uses no git word anywhere an editor reads", async () => {
    vi.stubGlobal(
      "fetch",
      server({
        [`${base}/review`]: aReview,
        [`POST ${base}/publishes`]: {
          id: "a-publish", state: "failed", changeCount: 2,
          startedAt: "2026-03-04T12:00:00Z",
          trouble: "We couldn't publish this time. Nothing changed on your live site, and your edits are safe.",
          steps: [{ title: "Putting it live", state: "failed" }],
        },
      }),
    );
    renderPage();

    await screen.findByText("Ready to publish 2 changes");
    for (const word of gitWords) {
      expect(readableText().toLowerCase()).not.toContain(word.toLowerCase());
    }

    await userEvent.click(screen.getByRole("button", { name: "Publish 2 changes" }));
    await screen.findByText(/Nothing changed on your live site/);

    for (const word of gitWords) {
      expect(readableText().toLowerCase()).not.toContain(word.toLowerCase());
    }
  });

  it("says you can keep editing", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/review`]: aReview }));
    renderPage();

    expect(await screen.findByText(/You can keep editing/)).toBeInTheDocument();
  });
});
