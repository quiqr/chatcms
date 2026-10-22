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
};

export function Avatar({ name, pending = false, className }: AvatarProps) {
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
