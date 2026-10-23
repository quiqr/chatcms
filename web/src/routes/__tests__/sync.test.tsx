import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { SyncPage } from "../SyncPage";

const base = "/api/organisations/technative/projects/korenaar";

function server(answers: Record<string, unknown>, sent?: (path: string, body: unknown) => void) {
  return vi.fn().mockImplementation((path: string, options?: RequestInit) => {
    const method = options?.method ?? "GET";
    if (method !== "GET") {
      sent?.(path, options?.body ? JSON.parse(String(options.body)) : null);
    }
    const key = method === "GET" ? path : `${method} ${path}`;
    return Promise.resolve({ ok: true, status: 200, json: async () => answers[key] ?? answers[path] ?? {} });
  });
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/p/korenaar/sync"]}>
      <Routes>
        <Route path="/p/:project/sync" element={<SyncPage organisation="technative" />} />
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

const states = {
  saved: { state: "saved", said: "All changes saved", unpublished: 0 },
  saving: { state: "saving", said: "Saving…", unpublished: 0 },
  offline: { state: "offline", said: "Offline · saved on this device", unpublished: 2 },
  arrived: {
    state: "arrived",
    said: "Someone else updated this site",
    unpublished: 0,
    arrivedFrom: ["Joris"],
  },
};

const asking = {
  state: "asking",
  said: "Two people changed the same thing",
  unpublished: 1,
  arrivedFrom: ["Joris"],
  conflicts: [
    {
      page: "home",
      title: "Home page",
      file: "content/_index.md",
      who: "Joris",
      fields: [
        {
          path: "summary",
          name: "Summary",
          mine: "We bake with apples.",
          theirs: "We bake with plums.",
          words: [
            { part: "kept", text: "We bake with " },
            { part: "gone", text: "apples." },
            { part: "added", text: "plums." },
          ],
        },
      ],
    },
  ],
};

describe("the sync status screen", () => {
  it("says each state in words a content creator reads", async () => {
    for (const [name, standing] of Object.entries(states)) {
      vi.stubGlobal("fetch", server({ [`${base}/standing`]: standing }));
      const { unmount } = renderPage();

      expect(await screen.findByText(standing.said)).toBeInTheDocument();
      if (name === "arrived") {
        expect(screen.getByText(/Joris updated this site/)).toBeInTheDocument();
      }
      unmount();
      vi.unstubAllGlobals();
    }
  });

  it("says how much is waiting to be published", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/standing`]: states.offline }));
    renderPage();

    expect(await screen.findByText(/2 changes are waiting to be published/)).toBeInTheDocument();
  });

  it("asks which to keep when two people changed the same field", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/standing`]: asking }));
    renderPage();

    expect(await screen.findByText(/Joris also changed this summary/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Keep mine" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Use Joris’s" })).toBeInTheDocument();
  });

  it("compares the two answers when asked", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/standing`]: asking }));
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: "Compare" }));

    expect(screen.getByText("apples.").tagName).toBe("DEL");
    expect(screen.getByText("plums.").tagName).toBe("INS");
  });

  it("sends what was chosen, field by field", async () => {
    const sent: Array<[string, unknown]> = [];
    vi.stubGlobal(
      "fetch",
      server(
        { [`${base}/standing`]: asking, [`POST ${base}/standing/answer`]: states.saved },
        (path, body) => sent.push([path, body]),
      ),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: "Use Joris’s" }));
    await userEvent.click(screen.getByRole("button", { name: "Save these answers" }));

    await waitFor(() =>
      expect(sent).toContainEqual([
        `${base}/standing/answer`,
        { answers: [{ file: "content/_index.md", fields: { summary: "theirs" } }] },
      ]),
    );
  });

  it("keeps mine unless something else is chosen", async () => {
    const sent: Array<[string, unknown]> = [];
    vi.stubGlobal(
      "fetch",
      server(
        { [`${base}/standing`]: asking, [`POST ${base}/standing/answer`]: states.saved },
        (path, body) => sent.push([path, body]),
      ),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: "Save these answers" }));

    await waitFor(() =>
      expect(sent).toContainEqual([
        `${base}/standing/answer`,
        { answers: [{ file: "content/_index.md", fields: {} }] },
      ]),
    );
  });

  it("offers one choice for a file it cannot show side by side", async () => {
    const whole = {
      ...asking,
      conflicts: [{ title: "Logo", file: "static/logo.png", who: "Joris", wholeFile: true }],
    };
    const sent: Array<[string, unknown]> = [];
    vi.stubGlobal(
      "fetch",
      server(
        { [`${base}/standing`]: whole, [`POST ${base}/standing/answer`]: states.saved },
        (path, body) => sent.push([path, body]),
      ),
    );
    renderPage();

    expect(await screen.findByText(/cannot show the two side by side/)).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Use theirs" }));
    await userEvent.click(screen.getByRole("button", { name: "Save these answers" }));

    await waitFor(() =>
      expect(sent).toContainEqual([
        `${base}/standing/answer`,
        { answers: [{ file: "static/logo.png", whole: "theirs" }] },
      ]),
    );
  });

  it("can start again instead of answering", async () => {
    const sent: Array<[string, unknown]> = [];
    vi.stubGlobal(
      "fetch",
      server(
        { [`${base}/standing`]: asking, [`POST ${base}/standing/abandon`]: states.saved },
        (path, body) => sent.push([path, body]),
      ),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: "Start again" }));
    await waitFor(() => expect(sent.some(([path]) => path.endsWith("/abandon"))).toBe(true));
  });

  it("shows technical details only when the server sent them", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/standing`]: states.saved }));
    const { unmount } = renderPage();
    await screen.findByText("All changes saved");
    expect(screen.queryByRole("button", { name: /Technical details/ })).not.toBeInTheDocument();
    unmount();
    vi.unstubAllGlobals();

    vi.stubGlobal(
      "fetch",
      server({
        [`${base}/standing`]: {
          ...states.saved,
          technical: { branch: "main", remote: "git@example.org:a/b.git", head: "abc123" },
        },
      }),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: /Technical details/ }));
    expect(screen.getByText("git@example.org:a/b.git")).toBeInTheDocument();
    expect(screen.getByText("abc123")).toBeInTheDocument();
  });

  it("uses no git word anywhere an editor reads", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/standing`]: asking }));
    renderPage();

    await screen.findByText(/Joris also changed this summary/);
    const words = (document.body.textContent ?? "").toLowerCase();
    for (const word of ["commit", "branch", "push", "pull", "merge", "repository", "git", "sha", "conflict"]) {
      expect(words).not.toContain(word);
    }
  });
});
