import { useState } from "react";
import { useParams } from "react-router";

import { Button } from "../components/ui/button";
import { demoProjects, demoViewer } from "../demo/content";
import { EditorShell, type EditorLayout } from "../shells/EditorShell";
import { PhoneShell } from "../shells/PhoneShell";
import { AssistantSlot } from "../shells/ProjectShell";
import { useIsPhone } from "../shells/useIsPhone";
import { PageEditor } from "./PageEditor";
import { Placeholder } from "./Placeholder";

export function EditorPage({ organisation }: { organisation: string }) {
  const { project: slug, page = "" } = useParams();
  const project =
    demoProjects.find((candidate) => candidate.slug === slug) ??
    demoProjects[0]!;
  const [layout, setLayout] = useState<EditorLayout>("split");
  const onAPhone = useIsPhone();

  if (!onAPhone) {
    return (
      <div>
        <EditorShell
          project={project}
          breadcrumb={page}
          layout={layout}
          onLayoutChange={setLayout}
          form={<PageEditor organisation={organisation} />}
          preview={<Placeholder screen="Preview" />}
        />
      </div>
    );
  }

  return (
    <PhoneShell
      project={project}
      viewer={demoViewer}
      bottomBar={
        <>
          <AssistantSlot />
          <Button className="flex-1" count={project.unpublished}>
            Publish
          </Button>
        </>
      }
    >
      <PageEditor organisation={organisation} />
    </PhoneShell>
  );
}
