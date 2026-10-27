import { useState } from "react";
import { useParams } from "react-router";

import { Button } from "../components/ui/button";
import { Segmented } from "../components/ui/segmented";
import { readProject } from "../lib/tenancy";
import { useAsync } from "../lib/useAsync";
import { EditorShell, type EditorLayout } from "../shells/EditorShell";
import { PhoneShell } from "../shells/PhoneShell";
import { AssistantSlot } from "../shells/ProjectShell";
import { useHere } from "../lib/whoami";
import { useIsPhone } from "../shells/useIsPhone";
import { PageEditor } from "./PageEditor";
import { Preview } from "./Preview";

export function EditorPage() {
  const { organisation, who } = useHere();
  const { project: slug, "*": page = "" } = useParams();
  const [loaded] = useAsync(() => readProject(organisation, slug ?? ""), [organisation, slug]);
  const [layout, setLayout] = useState<EditorLayout>("split");
  const [saves, setSaves] = useState(0);
  const [showing, setShowing] = useState<"edit" | "preview">("edit");
  const [opened, setOpened] = useState<{ title: string; section: string } | null>(null);
  const onAPhone = useIsPhone();

  if (loaded.state !== "ready") {
    return <p className="p-6 text-[15px] text-muted-foreground">Looking...</p>;
  }

  const project = loaded.value;

  if (!onAPhone) {
    return (
      <div>
        <EditorShell
          project={project}
          viewer={{ name: who.person.displayName || who.person.email }}
          breadcrumb={opened ? `${opened.section} · ${opened.title}` : page}
          layout={layout}
          onLayoutChange={setLayout}
          form={
            <PageEditor
              organisation={organisation}
              headed={false}
              onOpened={setOpened}
              onSaved={() => setSaves((count) => count + 1)}
            />
          }
          preview={<Preview organisation={organisation} project={slug ?? ""} page={page} showAgain={saves} />}
        />
      </div>
    );
  }

  return (
    <PhoneShell
      project={project}
      viewer={{ name: who.person.displayName || who.person.email, role: project.yourRole }}
      bottomBar={
        <>
          <AssistantSlot />
          <Button className="flex-1" count={project.unpublished}>
            Publish
          </Button>
        </>
      }
    >
      <div className="flex flex-col gap-4">
        <Segmented
          label="What to show"
          className="grid grid-cols-2"
          value={showing}
          options={[
            { value: "edit", label: "Edit" },
            { value: "preview", label: "Preview" },
          ]}
          onValueChange={(next) => setShowing(next as "edit" | "preview")}
        />
        {showing === "edit" ? (
          <PageEditor
            organisation={organisation}
            onOpened={setOpened}
            onSaved={() => setSaves((count) => count + 1)}
          />
        ) : (
          <div className="h-[70vh]">
            <Preview
              organisation={organisation}
              project={slug ?? ""}
              page={page}
              showAgain={saves}
            />
          </div>
        )}
      </div>
    </PhoneShell>
  );
}
