import { Image as ImageIcon, Trash2, Upload } from "lucide-react";
import { useRef, useState } from "react";
import { useParams } from "react-router";

import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Segmented } from "../components/ui/segmented";
import { useAsync } from "../lib/useAsync";
import { useHere } from "../lib/whoami";
import { plural, say } from "../lib/say";
import {
  addImage,
  describeImage,
  imageUrl,
  readImages,
  removeImage,
  type Image,
} from "../lib/tenancy";
import { whatWentWrong } from "../lib/api";

function inWords(bytes: number) {
  if (bytes >= 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  if (bytes >= 1024) return `${Math.round(bytes / 1024)} kB`;
  return `${bytes} bytes`;
}

function Details({
  organisation,
  project,
  image,
  picking,
  onDescribe,
  onRemove,
  onUse,
  onClose,
}: {
  organisation: string;
  project: string;
  image: Image;
  picking: boolean;
  onDescribe: (alt: string) => void;
  onRemove: (evenIfUsed: boolean) => void;
  onUse: (image: Image) => void;
  onClose: () => void;
}) {
  const [alt, setAlt] = useState(image.alt ?? "");
  const used = image.usedOn ?? [];

  return (
    <aside
      aria-label={`About ${image.name}`}
      className="flex flex-col gap-4 rounded-lg border border-border bg-card p-4"
    >
      <img
        src={imageUrl(organisation, project, image.name)}
        alt={image.alt ?? ""}
        className="max-h-64 w-full rounded-md object-contain"
      />
      <div className="flex flex-col gap-0.5">
        <p className="text-[15px] font-medium">{image.name}</p>
        <p className="text-sm text-subtle-foreground">
          {inWords(image.size)}
          {image.by ? ` · added by ${image.by}` : ""}
        </p>
      </div>

      <Field
        label={say("media.description-for-screen-readers")}
        hint={say("media.what-is-in-the-picture-for")}
        type="text"
        value={alt}
        onChange={(event) => setAlt(event.target.value)}
        onBlur={() => alt !== (image.alt ?? "") && onDescribe(alt)}
      />

      <div className="flex flex-col gap-1">
        <span className="text-sm font-medium">{say("media.used-on")}</span>
        {used.length === 0 ? (
          <p className="text-[15px] text-subtle-foreground">{say("media.not-used-anywhere-yet")}</p>
        ) : (
          <ul className="flex flex-col gap-0.5">
            {used.map((page) => (
              <li key={page} className="text-[15px]">
                {page}
              </li>
            ))}
          </ul>
        )}
      </div>

      <div className="flex flex-wrap gap-2">
        {picking && <Button onClick={() => onUse(image)}>{say("media.use-this-image")}</Button>}
        <Button variant="ghost" onClick={() => onRemove(false)}>
          <Trash2 aria-hidden="true" className="size-4" />
          {say("media.remove")}
        </Button>
        <Button variant="ghost" onClick={onClose}>
          {say("media.close")}
        </Button>
      </div>
    </aside>
  );
}

export function MediaPage({
  picking = false,
  onUse,
}: {
  picking?: boolean;
  onUse?: (image: Image) => void;
}) {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const [refresh, setRefresh] = useState(0);
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<"all" | "unused">("all");
  const [chosen, setChosen] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [trouble, setTrouble] = useState("");
  const [asking, setAsking] = useState<Image | null>(null);
  const picker = useRef<HTMLInputElement>(null);

  const [images] = useAsync<{ images: Image[] }>(
    () => readImages(organisation, project),
    [organisation, project, refresh],
  );

  async function add(files: FileList | null) {
    if (!files || files.length === 0) return;
    setBusy(true);
    setTrouble("");
    try {
      for (const file of Array.from(files)) {
        await addImage(organisation, project, file);
      }
      setRefresh(refresh + 1);
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error, "media.that-picture-could-not-be-added"));
    } finally {
      setBusy(false);
    }
  }

  async function describe(name: string, alt: string) {
    try {
      await describeImage(organisation, project, name, alt);
      setRefresh(refresh + 1);
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error, "media.that-description-was-not-saved"));
    }
  }

  async function take(image: Image, evenIfUsed: boolean) {
    if (!evenIfUsed && (image.usedOn ?? []).length > 0) {
      setAsking(image);
      return;
    }
    setBusy(true);
    try {
      await removeImage(organisation, project, image.name, evenIfUsed);
      setChosen(null);
      setAsking(null);
      setRefresh(refresh + 1);
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error, "media.that-picture-was-not-removed"));
    } finally {
      setBusy(false);
    }
  }

  if (images.state === "loading") {
    return <p className="text-[15px] text-subtle-foreground">{say("media.looking-at-your-pictures")}</p>;
  }
  if (images.state === "failed") {
    return <p className="text-[15px]">{images.message}</p>;
  }

  const looking = search.trim().toLowerCase();
  const shown = images.value.images.filter((image) => {
    if (filter === "unused" && (image.usedOn ?? []).length > 0) return false;
    if (looking === "") return true;
    return image.name.toLowerCase().includes(looking) || (image.alt ?? "").toLowerCase().includes(looking);
  });
  const selected = shown.find((image) => image.name === chosen) ?? null;

  return (
    <section className="flex flex-col gap-5">
      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      {asking && (
        <div
          role="dialog"
          aria-modal="true"
          aria-labelledby="remove-title"
          className="flex flex-col gap-4 rounded-lg border border-border bg-card p-4"
        >
          <h3 id="remove-title" className="text-base font-semibold">
            {say("media.remove-picture", { name: asking.name })}
          </h3>
          <p className="rounded-md bg-warning-soft p-3 text-[15px] text-warning-foreground">
            {say("media.it-is-on-pages", {
              pages: plural((asking.usedOn ?? []).length, "common.page-one", "common.page-many"),
            })}
          </p>
          <div className="flex flex-wrap gap-2">
            <Button onClick={() => take(asking, true)} disabled={busy}>
              {busy ? say("media.removing") : say("media.remove-it-anyway")}
            </Button>
            <Button variant="ghost" onClick={() => setAsking(null)} disabled={busy}>
              {say("common.keep-it")}
            </Button>
          </div>
        </div>
      )}

      <div className="flex flex-wrap items-end gap-3">
        <div className="min-w-[200px] flex-1">
          <Field
            label={say("media.search")}
            type="search"
            value={search}
            placeholder={say("media.find-a-picture")}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <Segmented
          label={say("media.which-pictures-to-show")}
          value={filter}
          options={[
            { value: "all", label: `All ${images.value.images.length}` },
            { value: "unused", label: say("media.not-used-anywhere") },
          ]}
          onValueChange={(next) => setFilter(next as "all" | "unused")}
        />
        <Button onClick={() => picker.current?.click()} disabled={busy}>
          <Upload aria-hidden="true" className="size-4" />
          {busy ? say("media.adding") : say("media.add-pictures")}
        </Button>
        <input
          ref={picker}
          type="file"
          accept="image/*"
          multiple
          aria-label={say("media.add-pictures")}
          onChange={(event) => add(event.target.files)}
          className="sr-only"
        />
      </div>

      <div
        onDragOver={(event) => event.preventDefault()}
        onDrop={(event) => {
          event.preventDefault();
          add(event.dataTransfer.files);
        }}
        className="rounded-lg border border-dashed border-input p-4"
      >
        {shown.length === 0 ? (
          <p className="text-center text-[15px] text-subtle-foreground">
            {images.value.images.length === 0
              ? say("media.no-pictures-yet-drop-some-here")
              : say("media.nothing-here-matches-that")}
          </p>
        ) : (
          <ul className="grid gap-3 [grid-template-columns:repeat(auto-fill,minmax(150px,1fr))]">
            {shown.map((image) => (
              <li key={image.name}>
                <button
                  type="button"
                  onClick={() => setChosen(image.name === chosen ? null : image.name)}
                  aria-pressed={image.name === chosen}
                  className={
                    "flex w-full flex-col gap-1 rounded-md border p-1 text-left " +
                    (image.name === chosen ? "border-primary" : "border-transparent hover:border-border")
                  }
                >
                  <img
                    src={imageUrl(organisation, project, image.name, true)}
                    alt={image.alt ?? ""}
                    loading="lazy"
                    className="aspect-square w-full rounded bg-muted object-cover"
                  />
                  <span className="truncate text-sm">{image.name}</span>
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      {selected && (
        <Details
          organisation={organisation}
          project={project}
          image={selected}
          picking={picking}
          onDescribe={(alt) => describe(selected.name, alt)}
          onRemove={(evenIfUsed) => take(selected, evenIfUsed)}
          onUse={(image) => onUse?.(image)}
          onClose={() => setChosen(null)}
        />
      )}

      {shown.length === 0 && images.value.images.length === 0 && (
        <p className="flex items-center gap-2 text-sm text-subtle-foreground">
          <ImageIcon aria-hidden="true" className="size-4" />
          {say("media.pictures-live-in-your-site-so")}
        </p>
      )}
    </section>
  );
}
