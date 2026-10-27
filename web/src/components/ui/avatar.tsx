import { useState } from "react";

import { cn } from "../../lib/cn";

export function initialsOf(name: string) {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  if (parts.length === 1) return parts[0]!.slice(0, 2).toUpperCase();
  return (parts[0]![0]! + parts[parts.length - 1]![0]!).toUpperCase();
}

export type AvatarProps = {
  name: string;
  pending?: boolean;
  className?: string;
  /** personId, when given, is whose picture to show instead of initials. */
  personId?: string;
  /**
   * changed is bumped when the picture has just been replaced or removed.
   * The browser is told it may keep a picture for a minute, which is right
   * everywhere it is shown beside a name and wrong on the screen that has
   * just changed it.
   */
  changed?: number;
};

export function Avatar({ name, pending = false, className, personId, changed = 0 }: AvatarProps) {
  const [hasPicture, setHasPicture] = useState(Boolean(personId));

  // Initials are what somebody has until they add a picture, and what they
  // go back to if it cannot be loaded. The image says whether it worked
  // rather than this asking first.
  if (personId && hasPicture) {
    return (
      <img
        src={`/api/people/${personId}/picture${changed > 0 ? `?changed=${changed}` : ""}`}
        alt={name}
        onError={() => setHasPicture(false)}
        className={cn("inline-block size-8 shrink-0 rounded-full object-cover", className)}
      />
    );
  }

  return (
    <span
      role="img"
      aria-label={name}
      className={cn(
        "inline-flex size-8 shrink-0 items-center justify-center rounded-full bg-primary-soft text-xs font-semibold text-primary-soft-foreground",
        pending && "border border-dashed border-input bg-transparent text-subtle-foreground",
        className,
      )}
    >
      {initialsOf(name)}
    </span>
  );
}

export type TileProps = {
  name: string;
  className?: string;
};

export function Tile({ name, className }: TileProps) {
  return (
    <span
      aria-hidden="true"
      className={cn(
        "inline-flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-sm font-semibold text-primary-soft-foreground",
        className,
      )}
    >
      {name.trim().slice(0, 1).toUpperCase()}
    </span>
  );
}
