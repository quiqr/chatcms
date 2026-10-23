import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { PageEditor } from "../PageEditor";


const aPage = {
  key: "home",
  title: "Home page",
  section: "Pages",
  hasBody: true,
  body: "The body of the page.\n",
  fields: [
    { path: "title", type: "string", title: "Page title", value: "Welcome" },
    {
      path: "intro",
      type: "string",
      title: "Introduction",
      value: "We sell apples.",
      config: { multiLine: true },
    },
    { path: "published", type: "boolean", title: "Published", value: true },
    {
      path: "colour",
      type: "select",
      title: "Colour",
      value: "red",
      options: [
        { value: "red", label: "Red" },
        { value: "blue", label: "Blue" },
      ],
    },
    { path: "when", type: "date", title: "Date", value: "2026-03-01T09:00:00Z" },
    { path: "tags", type: "leaf-array", title: "Tags", value: ["bread", "cake"] },
    {
      path: "links",
      type: "accordion",
      title: "Links",
      value: [{ label: "Shop", url: "/shop/" }],
      fields: [
        { path: "label", type: "string", title: "Label" },
        { path: "url", type: "string", title: "Address" },
      ],
    },
    {
      path: "seo",
      type: "section",
      title: "Search",
      fields: [{ path: "seo.description", type: "string", title: "Description", value: "A bakery." }],
    },
    { path: "mystery", type: "eisenhouwer", title: "Mystery", value: "something", unknown: true },
  ],
};

function server(sent?: (body: unknown) => void, answer: unknown = aPage) {
  return vi.fn().mockImplementation((_path: string, options?: RequestInit) => {
    if (options?.body) sent?.(JSON.parse(String(options.body)));
    return Promise.resolve({ ok: true, status: 200, json: async () => answer });
  });
}

function renderEditor() {
  return render(
    <MemoryRouter initialEntries={["/p/korenaar/content/home"]}>
      <Routes>
        <Route path="/p/:project/content/*" element={<PageEditor organisation="technative" />} />
      </Routes>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
  vi.useRealTimers();
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("the page editor", () => {
  it("draws a control for each kind of field the model declares", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    expect(await screen.findByLabelText("Page title")).toHaveValue("Welcome");
    expect(screen.getByLabelText("Introduction")).toHaveValue("We sell apples.");
    expect(screen.getByLabelText("Published")).toBeChecked();
    expect(screen.getByLabelText("Colour")).toHaveValue("red");
    expect(screen.getByLabelText("Date")).toHaveValue("2026-03-01");
    expect(screen.getByLabelText("Tags 1")).toHaveValue("bread");
    expect(screen.getByLabelText("Page text")).toHaveValue("The body of the page.\n");
  });

  it("shows a group with the fields inside it", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    await screen.findByLabelText("Page title");
    expect(screen.getByRole("button", { name: "Search" })).toBeInTheDocument();
    expect(screen.getByLabelText("Description")).toHaveValue("A bakery.");
  });

  it("shows a field it cannot draw as plain text, and says so", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    expect(await screen.findByLabelText("Mystery")).toHaveValue("something");
    expect(screen.getByText(/shown as plain text for now/)).toBeInTheDocument();
  });

  it("saves what was typed, as one save rather than one per letter", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", server((body) => sent.push(body)));
    renderEditor();

    const title = await screen.findByLabelText("Page title");
    await userEvent.clear(title);
    await userEvent.type(title, "Welcome home");

    await waitFor(() => expect(sent.length).toBeGreaterThan(0), { timeout: 4000 });
    expect(sent).toHaveLength(1);
    expect(sent[0]).toEqual({ values: { title: "Welcome home" } });
  });

  it("says when it is saving and when it is done", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    const title = await screen.findByLabelText("Page title");
    expect(screen.getByText("All changes saved")).toBeInTheDocument();

    await userEvent.type(title, "!");
    await waitFor(() => expect(screen.getByText("All changes saved")).toBeInTheDocument(), {
      timeout: 4000,
    });
  });

  it("says so when a save does not work", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation((_path: string, options?: RequestInit) => {
        if (options?.method === "PUT") {
          return Promise.resolve({
            ok: false,
            status: 409,
            json: async () => ({ error: { code: "refused", message: "That could not be saved." } }),
          });
        }
        return Promise.resolve({ ok: true, status: 200, json: async () => aPage });
      }),
    );
    renderEditor();

    await userEvent.type(await screen.findByLabelText("Page title"), "!");

    expect(await screen.findByText(/keep trying/, {}, { timeout: 4000 })).toBeInTheDocument();
  });

  it("adds and removes an item in a list of values", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", server((body) => sent.push(body)));
    renderEditor();

    await screen.findByLabelText("Tags 1");
    await userEvent.click(screen.getByRole("button", { name: /Add tags/i }));
    expect(screen.getByLabelText("Tags 3")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Remove Tags 1" }));
    await waitFor(() => expect(screen.getByLabelText("Tags 1")).toHaveValue("cake"));
  });

  it("opens an item in a list of items and edits a field inside it", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", server((body) => sent.push(body)));
    renderEditor();

    await screen.findByLabelText("Page title");
    expect(screen.getByLabelText("Label")).toHaveValue("Shop");

    await userEvent.clear(screen.getByLabelText("Label"));
    await userEvent.type(screen.getByLabelText("Label"), "Our shop");

    await waitFor(() => expect(sent.length).toBeGreaterThan(0), { timeout: 4000 });
    expect(sent[0]).toEqual({ values: { links: [{ label: "Our shop", url: "/shop/" }] } });
  });

  it("writes Markdown from the toolbar", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    const text = await screen.findByLabelText("Page text");
    await userEvent.clear(text);
    await userEvent.type(text, "plain");
    (text as HTMLTextAreaElement).setSelectionRange(0, 5);

    await userEvent.click(screen.getByRole("button", { name: "Bold" }));

    await waitFor(() => expect(text).toHaveValue("**plain**"));
  });

  it("can show the Markdown source instead", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    await screen.findByLabelText("Page text");
    expect(screen.getByRole("toolbar", { name: "Text formatting" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("radio", { name: "Markdown" }));
    expect(screen.queryByRole("toolbar", { name: "Text formatting" })).not.toBeInTheDocument();
  });

  it("shows a page's text but does not edit it when the model gives it no field", async () => {
    vi.stubGlobal("fetch", server(undefined, { ...aPage, hasBody: false }));
    renderEditor();

    await screen.findByLabelText("Page title");
    expect(screen.queryByLabelText("Page text")).not.toBeInTheDocument();
    expect(screen.getByText(/not set up to be edited here/)).toBeInTheDocument();
    expect(screen.getByText(/The body of the page/)).toBeInTheDocument();
  });

  it("has no Save button anywhere", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    await screen.findByLabelText("Page title");
    expect(screen.queryByRole("button", { name: /^Save/i })).not.toBeInTheDocument();
  });

  it("uses no git word anywhere an editor reads", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    await screen.findByLabelText("Page title");
    const words = (document.body.textContent ?? "").toLowerCase();
    for (const word of ["commit", "branch", "push", "merge", "repository", "git", "sha", "head"]) {
      expect(words).not.toContain(word);
    }
  });
});
