import { useState } from "react";
import { useNavigate, useParams } from "react-router";

import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { useAsync } from "../lib/useAsync";
import { changeProject, readProject, removeProject } from "../lib/tenancy";
import { useHere } from "../lib/whoami";
import { canSeeTechnicalDetail } from "../shells/roles";
import { KeyPanel, PublishingPanel } from "./setup/Panels";

function Panel({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="rounded-xl border border-border bg-card p-4">
      <h2 className="text-[15px] font-semibold">{title}</h2>
      <div className="mt-3">{children}</div>
    </section>
  );
}

export function ProjectSettingsPage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const goTo = useNavigate();

  const [loaded, again] = useAsync(() => readProject(organisation, project), [organisation, project]);

  const [name, setName] = useState<string | null>(null);
  const [domain, setDomain] = useState<string | null>(null);
  const [trouble, setTrouble] = useState("");
  const [saved, setSaved] = useState(false);
  const [busy, setBusy] = useState(false);
  const [removing, setRemoving] = useState(false);

  if (loaded.state === "loading") {
    return <p className="text-[15px] text-muted-foreground">Looking...</p>;
  }
  if (loaded.state === "failed") {
    return (
      <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
        {loaded.message}
      </p>
    );
  }

  const here = loaded.value;
  const shownName = name ?? here.name;
  const shownDomain = domain ?? here.domain;
  const mayChange = here.youMay.includes("change-settings");
  const mayRemove = here.youMay.includes("delete-project");

  if (!mayChange) {
    return (
      <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
        You are not allowed to change this project's settings.
      </p>
    );
  }

  async function save(event: React.FormEvent) {
    event.preventDefault();
    setBusy(true);
    setTrouble("");
    setSaved(false);
    try {
      await changeProject(organisation, project, {
        name: shownName.trim(),
        domain: shownDomain.trim(),
      });
      again();
      setSaved(true);
    } catch (error: unknown) {
      setTrouble(error instanceof Error ? error.message : "That did not work.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex max-w-[640px] flex-col gap-6">
      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      <Panel title="What it is called">
        <form onSubmit={save} className="flex flex-col gap-4">
          <Field
            label="Name"
            value={shownName}
            onChange={(event) => setName(event.target.value)}
            required
          />
          <Field
            label="Address shown to customers"
            value={shownDomain}
            onChange={(event) => setDomain(event.target.value)}
            hint="Where the site can be visited, for your own reference."
          />

          <p className="text-[13px] text-subtle-foreground">
            Its address here stays /o/{organisation}/p/{here.slug}, so links already sent keep
            working.
          </p>

          {saved && (
            <p role="status" className="text-[15px] text-muted-foreground">
              Saved.
            </p>
          )}

          <div>
            <Button type="submit" disabled={busy || shownName.trim() === ""}>
              {busy ? "Saving..." : "Save"}
            </Button>
          </div>
        </form>
      </Panel>

      {canSeeTechnicalDetail(here.yourRole) && (
        <>
          <Panel title="Where the site lives">
            <KeyPanel organisation={organisation} project={project} onTrouble={setTrouble} />
          </Panel>

          <Panel title="How it is published">
            <PublishingPanel organisation={organisation} project={project} onTrouble={setTrouble} />
          </Panel>
        </>
      )}

      {mayRemove && (
        <Panel title="Remove this project">
          {removing ? (
            <div className="flex flex-col gap-3">
              <p className="text-[15px]">
                This removes <strong>{here.name}</strong> from ChatCMS: its settings, who may use it,
                and its history here.
              </p>
              <p className="text-[15px]">
                The site itself is not touched. Everything published stays exactly where it is, and
                nothing is deleted from where the site lives.
              </p>

              <div className="flex flex-wrap gap-2">
                <Button
                  variant="destructive"
                  disabled={busy}
                  onClick={() => {
                    setBusy(true);
                    void removeProject(organisation, project)
                      .then(() => goTo(`/o/${organisation}/projects`))
                      .catch((error: unknown) => {
                        setTrouble(error instanceof Error ? error.message : "That did not work.");
                        setBusy(false);
                      });
                  }}
                >
                  Yes, remove it
                </Button>
                <Button variant="outline" onClick={() => setRemoving(false)}>
                  Keep it
                </Button>
              </div>
            </div>
          ) : (
            <Button variant="outline" onClick={() => setRemoving(true)}>
              Remove this project...
            </Button>
          )}
        </Panel>
      )}
    </div>
  );
}
