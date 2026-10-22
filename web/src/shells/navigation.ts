import {
  FileText,
  History,
  Image,
  SlidersHorizontal,
  Upload,
  Users,
  Workflow,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

export type NavEntry = {
  key: string;
  label: string;
  to: string;
  icon: LucideIcon;
  soon?: boolean;
};

export function siteNavigation(projectSlug: string): NavEntry[] {
  return [
    { key: "content", label: "Content", to: `/p/${projectSlug}/content`, icon: FileText },
    { key: "media", label: "Media", to: `/p/${projectSlug}/media`, icon: Image },
    { key: "publish", label: "Publish", to: `/p/${projectSlug}/publish`, icon: Upload },
    { key: "history", label: "History", to: `/p/${projectSlug}/history`, icon: History },
  ];
}

export function setupNavigation(projectSlug: string): NavEntry[] {
  return [
    { key: "people", label: "People & roles", to: `/p/${projectSlug}/people`, icon: Users },
    { key: "settings", label: "Project settings", to: `/p/${projectSlug}/settings`, icon: SlidersHorizontal },
    { key: "workflows", label: "Workflows", to: `/p/${projectSlug}/workflows`, icon: Workflow, soon: true },
  ];
}
