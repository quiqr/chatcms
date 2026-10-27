import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";

import { Inside } from "./here";import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { MediaPage } from "../MediaPage";

const base = "/api/organisations/technative/projects/korenaar";

const theImages = {
  images: [
    { name: "logo.png", path: "static/images/logo.png", size: 2048, alt: "The sign", usedOn: ["home"] },
    { name: "bread.png", path: "static/images/bread.png", size: 1048576, by: "Anna", usedOn: [] },
  ],
};

function server(answers: Record<string, unknown>, sent?: (path: string, method: string) => void) {
  return vi.fn().mockImplementation((path: string, options?: RequestInit) => {
    const method = options?.method ?? "GET";
    if (method !== "GET") sent?.(path, method);
    const key = method === "GET" ? path : `${method} ${path}`;
    return Promise.resolve({
      ok: true,
      status: method === "DELETE" ? 204 : 200,
      json: async () => answers[key] ?? answers[path] ?? {},
    });
  });
}

function renderPage(props: { picking?: boolean; onUse?: (image: unknown) => void } = {}) {
  return render(
    <Inside><MemoryRouter initialEntries={["/o/technative/p/korenaar/media"]}>
      <Routes>
        <Route path="/o/:organisation/p/:project/media" element={<MediaPage {...props} />} />
      </Routes>
    </MemoryRouter></Inside>,
  );
}

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("the media library", () => {
  it("shows every picture with its name", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/images`]: theImages }));
    renderPage();

    expect(await screen.findByText("logo.png")).toBeInTheDocument();
    expect(screen.getByText("bread.png")).toBeInTheDocument();
  });

  it("can show only the pictures nothing uses", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/images`]: theImages }));
    renderPage();

    await screen.findByText("logo.png");
    await userEvent.click(screen.getByRole("radio", { name: "Not used anywhere" }));

    expect(screen.getByText("bread.png")).toBeInTheDocument();
    expect(screen.queryByText("logo.png")).not.toBeInTheDocument();
  });

  it("opens a picture and says what uses it and how large it is", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/images`]: theImages }));
    renderPage();

    await userEvent.click(await screen.findByText("logo.png"));

    const details = screen.getByRole("complementary", { name: "About logo.png" });
    expect(details).toHaveTextContent("2 kB");
    expect(details).toHaveTextContent("home");
    expect(within(details).getByLabelText("Description for screen readers")).toHaveValue("The sign");
  });

  it("says when a picture is used nowhere", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/images`]: theImages }));
    renderPage();

    await userEvent.click(await screen.findByText("bread.png"));

    const details = screen.getByRole("complementary", { name: "About bread.png" });
    expect(details).toHaveTextContent("Not used anywhere yet");
    expect(details).toHaveTextContent("1.0 MB");
  });

  it("warns before removing a picture something still uses", async () => {
    const sent: Array<[string, string]> = [];
    vi.stubGlobal("fetch", server({ [`${base}/images`]: theImages }, (path, method) => sent.push([path, method])));
    renderPage();

    await userEvent.click(await screen.findByText("logo.png"));
    await userEvent.click(screen.getByRole("button", { name: /Remove/ }));

    const dialog = await screen.findByRole("dialog");
    expect(dialog).toHaveTextContent("gap where it was");
    expect(sent).toHaveLength(0);

    await userEvent.click(within(dialog).getByRole("button", { name: "Remove it anyway" }));
    await waitFor(() => expect(sent).toContainEqual([`${base}/images/logo.png?evenIfUsed=true`, "DELETE"]));
  });

  it("removes a picture nothing uses without asking", async () => {
    const sent: Array<[string, string]> = [];
    vi.stubGlobal("fetch", server({ [`${base}/images`]: theImages }, (path, method) => sent.push([path, method])));
    renderPage();

    await userEvent.click(await screen.findByText("bread.png"));
    await userEvent.click(screen.getByRole("button", { name: /Remove/ }));

    await waitFor(() => expect(sent).toContainEqual([`${base}/images/bread.png`, "DELETE"]));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("writes a description when the field is left", async () => {
    const sent: Array<[string, string]> = [];
    vi.stubGlobal(
      "fetch",
      server({ [`${base}/images`]: theImages, [`PUT ${base}/images/bread.png`]: {} }, (path, method) =>
        sent.push([path, method]),
      ),
    );
    renderPage();

    await userEvent.click(await screen.findByText("bread.png"));
    const field = screen.getByLabelText("Description for screen readers");
    await userEvent.type(field, "A loaf on a board");
    await userEvent.tab();

    await waitFor(() => expect(sent).toContainEqual([`${base}/images/bread.png`, "PUT"]));
  });

  it("offers to use a picture only when it was opened to pick one", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/images`]: theImages }));
    const { unmount } = renderPage();

    await userEvent.click(await screen.findByText("logo.png"));
    expect(screen.queryByRole("button", { name: "Use this image" })).not.toBeInTheDocument();
    unmount();

    const used: unknown[] = [];
    renderPage({ picking: true, onUse: (image) => used.push(image) });
    await userEvent.click(await screen.findByText("logo.png"));
    await userEvent.click(screen.getByRole("button", { name: "Use this image" }));

    expect(used).toHaveLength(1);
  });

  it("says what to do when a site has no pictures", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/images`]: { images: [] } }));
    renderPage();

    expect(await screen.findByText(/No pictures yet/)).toBeInTheDocument();
  });

  it("uses no git word anywhere an editor reads", async () => {
    vi.stubGlobal("fetch", server({ [`${base}/images`]: theImages }));
    renderPage();

    await screen.findByText("logo.png");
    const words = (document.body.textContent ?? "").toLowerCase();
    for (const word of ["commit", "branch", "push", "merge", "repository", "git", "sha"]) {
      expect(words).not.toContain(word);
    }
  });
});
