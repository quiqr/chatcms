import { useState } from "react";
import { useNavigate, useParams } from "react-router";

import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { useAsync } from "../lib/useAsync";
import { changeProject, readProject, removeProject } from "../lib/tenancy";
import { useHere } from "../lib/whoami";
import { canSeeTechnicalDetail } from "../shells/roles";
import { KeyPanel, PublishingPanel } from "./setup/Panels";
import { say } from "../lib/say";
import { whatWentWrong } from "../lib/api";

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
    return <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>;
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
        {say("project-settings.you-are-not-allowed-to-change")}
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
      setTrouble(whatWentWrong(error));
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

      <Panel title={say("project-settings.what-it-is-called")}>
        <form onSubmit={save} className="flex flex-col gap-4">
          <Field
            label={say("common.name")}
            value={shownName}
            onChange={(event) => setName(event.target.value)}
            required
          />
          <Field
            label={say("project-settings.address-shown-to-customers")}
            value={shownDomain}
            onChange={(event) => setDomain(event.target.value)}
            hint={say("project-settings.where-the-site-can-be-visited")}
          />

          <p className="text-[13px] text-subtle-foreground">
            {say("project-settings.its-address-stays", {
              organisation,
              project: here.slug,
            })}
          </p>

          {saved && (
            <p role="status" className="text-[15px] text-muted-foreground">
              {say("common.saved")}
            </p>
          )}

          <div>
            <Button type="submit" disabled={busy || shownName.trim() === ""}>
              {busy ? say("common.saving") : say("project-settings.save")}
            </Button>
          </div>
        </form>
      </Panel>

      {canSeeTechnicalDetail(here.yourRole) && (
        <>
          <Panel title={say("project-settings.where-the-site-lives")}>
            <KeyPanel organisation={organisation} project={project} onTrouble={setTrouble} />
          </Panel>

          <Panel title={say("project-settings.how-it-is-published")}>
            <PublishingPanel organisation={organisation} project={project} onTrouble={setTrouble} />
          </Panel>
        </>
      )}

      {mayRemove && (
        <Panel title={say("project-settings.remove-this-project")}>
          {removing ? (
            <div className="flex flex-col gap-3">
              <p className="text-[15px]">
                {say("project-settings.this-removes-it", { name: here.name })}
              </p>
              <p className="text-[15px]">
                {say("project-settings.the-site-itself-is-not-touched")}
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
                        setTrouble(whatWentWrong(error));
                        setBusy(false);
                      });
                  }}
                >
                  {say("common.yes-remove-it")}
                </Button>
                <Button variant="outline" onClick={() => setRemoving(false)}>
                  {say("common.keep-it")}
                </Button>
              </div>
            </div>
          ) : (
            <Button variant="outline" onClick={() => setRemoving(true)}>
              {say("project-settings.remove-this-project-2")}
            </Button>
          )}
        </Panel>
      )}
    </div>
  );
}
