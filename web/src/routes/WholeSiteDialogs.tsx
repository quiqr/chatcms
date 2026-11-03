import { Ban } from "lucide-react";
import { useState } from "react";

import { Button } from "../components/ui/button";
import { Checkbox } from "../components/ui/field";
import { Field } from "../components/ui/field";
import { whyThereIsNoRoom } from "../components/ui/usage";
import { say } from "../lib/say";
import type { OrganisationArea, Project } from "../lib/tenancy";
import type { WhoAmI } from "../lib/whoami";

function slugFrom(name: string) {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

/**
 * Duplicating a site. The name starts as the original's and is refused from
 * the first keystroke, which is the design's way of saying "think about
 * this" rather than making somebody delete a word before they can type.
 *
 * chatcms-audience: implementer
 */
export function Duplicating({
  area,
  project,
  busy,
  onYes,
  onCancel,
}: {
  area: OrganisationArea;
  project: Project;
  busy: boolean;
  onYes: (body: { slug: string; name: string; withPeople: boolean; withPublishing: boolean }) => void;
  onCancel: () => void;
}) {
  const [name, setName] = useState(project.name);
  const [withPeople, setWithPeople] = useState(false);
  const [withPublishing, setWithPublishing] = useState(false);

  const taken = name.trim() === project.name.trim();
  const noRoom = whyThereIsNoRoom(area.usage.projects, "sites");

  return (
    <section
      aria-labelledby="duplicate-title"
      className="flex flex-col gap-4 rounded-lg border border-border bg-card p-4"
    >
      <h3 id="duplicate-title" className="text-base font-semibold">
        {say("copy.duplicate-site", { name: project.name })}
      </h3>
      <p className="text-[15px] text-muted-foreground">
        {say("copy.makes-a-separate-copy", { organisation: area.name })}
      </p>

      <Field
        label={say("copy.name-of-the-copy")}
        value={name}
        onChange={(event) => setName(event.target.value)}
        required
        trouble={taken ? say("copy.that-name-is-taken") : undefined}
      />

      <fieldset className="flex flex-col gap-2">
        <legend className="text-sm font-medium">{say("copy.what-to-copy")}</legend>

        <p className="text-[15px] text-muted-foreground">
          {say("copy.pages-media-and-model")} · {say("copy.always")}
        </p>

        <Checkbox
          label={say("copy.people-and-their-roles")}
          checked={withPeople}
          onChange={(event) => setWithPeople(event.target.checked)}
        />
        <Checkbox
          label={say("copy.publishing-settings")}
          checked={withPublishing}
          onChange={(event) => setWithPublishing(event.target.checked)}
        />
        <p className="text-[13px] text-subtle-foreground">{say("copy.off-by-default")}</p>
      </fieldset>

      {noRoom && (
        <p className="flex items-start gap-2 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          <Ban aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
          {noRoom}
        </p>
      )}

      <div className="flex flex-wrap gap-2">
        <Button
          disabled={busy || taken || name.trim() === "" || noRoom !== null}
          onClick={() =>
            onYes({
              slug: slugFrom(name),
              name: name.trim(),
              withPeople,
              withPublishing,
            })
          }
        >
          {busy ? say("copy.duplicating") : say("copy.duplicate-it")}
        </Button>
        <Button variant="outline" onClick={onCancel}>
          {say("content.cancel")}
        </Button>
      </div>
    </section>
  );
}

/**
 * Moving a site. Only organisations somebody administers can be chosen; the
 * rest are shown with the reason, because "it is not in the list" leaves
 * somebody wondering whether the product forgot it.
 *
 * chatcms-audience: implementer
 */
export function Moving({
  who,
  from,
  project,
  busy,
  onYes,
  onCancel,
}: {
  who: WhoAmI;
  from: string;
  project: Project;
  busy: boolean;
  onYes: (to: string) => void;
  onCancel: () => void;
}) {
  const elsewhere = who.organisations.filter((one) => one.slug !== from);
  const allowed = elsewhere.filter((one) => one.youAdminister);
  const [chosen, setChosen] = useState(allowed[0]?.slug ?? "");

  return (
    <section
      aria-labelledby="move-title"
      className="flex flex-col gap-4 rounded-lg border border-border bg-card p-4"
    >
      <h3 id="move-title" className="text-base font-semibold">
        {say("move.move-site", { name: project.name })}
      </h3>

      {allowed.length === 0 ? (
        <p className="text-[15px] text-muted-foreground">{say("move.nowhere-to-move-it")}</p>
      ) : (
        <fieldset className="flex flex-col gap-2">
          <legend className="text-sm font-medium">{say("move.move-to")}</legend>

          {elsewhere.map((one) => (
            <label
              key={one.slug}
              className={`flex min-h-11 items-start gap-2 rounded-md border p-3 ${
                one.youAdminister
                  ? "border-input hover:bg-muted"
                  : "border-border-subtle text-subtle-foreground"
              }`}
            >
              {/* chatcms-target: the label around this one carries the hit area */}
              <input
                type="radio"
                name="move-to"
                className="mt-1 size-4"
                value={one.slug}
                checked={chosen === one.slug}
                disabled={!one.youAdminister}
                onChange={() => setChosen(one.slug)}
              />
              <span className="flex flex-col">
                <span className="text-[15px]">{one.name}</span>
                <span className="text-[13px]">
                  {one.youAdminister ? say("move.you-administer-it") : say("move.you-only-belong")}
                </span>
              </span>
            </label>
          ))}
        </fieldset>
      )}

      <div className="flex flex-col gap-1">
        <h4 className="text-sm font-medium">{say("move.what-happens")}</h4>
        <ul className="flex flex-col gap-1 text-[15px] text-muted-foreground">
          <li>{say("move.the-website-moves-along")}</li>
          <li>{say("move.the-people-come-along")}</li>
          <li>{say("move.the-old-admins-lose-it")}</li>
        </ul>
      </div>

      <div className="flex flex-wrap gap-2">
        <Button disabled={busy || chosen === ""} onClick={() => onYes(chosen)}>
          {busy ? say("move.moving") : say("move.move-it")}
        </Button>
        <Button variant="outline" onClick={onCancel}>
          {say("content.cancel")}
        </Button>
      </div>
    </section>
  );
}
