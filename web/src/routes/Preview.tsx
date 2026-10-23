import { Monitor, Smartphone } from "lucide-react";
import { useState } from "react";

import { Segmented } from "../components/ui/segmented";

/**
 * Preview shows the site as a reader will see it. It is served through the
 * app's own address, so somebody who may not see the project cannot see its
 * unpublished work either.
 */
export function Preview({
  organisation,
  project,
  page,
  showAgain = 0,
}: {
  organisation: string;
  project: string;
  page?: string;
  /**
   * showAgain is bumped by whatever knows the site changed. Hugo rebuilds by
   * itself, but the page already in the frame cannot be told so: its live
   * reload talks to Hugo directly, and Hugo is not reachable from a browser
   * on purpose.
   */
  showAgain?: number;
}) {
  const [device, setDevice] = useState<"desktop" | "phone">("desktop");
  const [reloads, setReloads] = useState(0);

  const at = `/preview/${organisation}/${project}/${pathOf(page)}`;

  return (
    <div className="flex h-full flex-col gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="truncate rounded-full bg-muted px-3 py-1 text-[13px] text-subtle-foreground">
          Preview · {at}
        </span>
        <div className="flex items-center gap-2">
          <Segmented
            label="How wide to show the site"
            value={device}
            options={[
              { value: "desktop", label: "Desktop" },
              { value: "phone", label: "Phone" },
            ]}
            onValueChange={(next) => setDevice(next as "desktop" | "phone")}
          />
          <button
            type="button"
            onClick={() => setReloads(reloads + 1)}
            className="flex min-h-11 items-center px-2 text-sm font-medium text-primary"
          >
            Refresh
          </button>
        </div>
      </div>

      <div className="flex flex-1 justify-center overflow-hidden rounded-lg border border-border bg-card">
        <iframe
          key={`${reloads}-${showAgain}`}
          title="Your site"
          src={at}
          className={
            "h-full border-0 bg-card " + (device === "phone" ? "w-[390px]" : "w-full")
          }
        />
      </div>

      <p className="flex items-center gap-2 text-[13px] text-subtle-foreground">
        {device === "phone" ? (
          <Smartphone aria-hidden="true" className="size-4" />
        ) : (
          <Monitor aria-hidden="true" className="size-4" />
        )}
        This is your site as a visitor will see it. Changes show within a few seconds.
      </p>
    </div>
  );
}

/**
 * pathOf turns a page's key into the address Hugo serves it at: a single is
 * the site's front page unless it says otherwise, and an item of a collection
 * sits under its collection.
 */
function pathOf(page?: string) {
  if (!page || page === "home") return "";
  return page.endsWith("/") ? page : page + "/";
}
