import { useState } from "react";
import { useParams } from "react-router";

import { Button } from "../components/ui/button";
import { demoProjects, demoViewer } from "../demo/content";
import { EditorShell, type EditorLayout } from "../shells/EditorShell";
import { PhoneShell } from "../shells/PhoneShell";
import { AssistantSlot } from "../shells/ProjectShell";
import { Placeholder } from "./Placeholder";

export function EditorPage() {
  const { project: slug } = useParams();
  const project = demoProjects.find((candidate) => candidate.slug === slug) ?? demoProjects[0]!;
  const [layout, setLayout] = useState<EditorLayout>("split");

  return (
    <>
      <div className="hidden sm:block">
        <EditorShell
          project={project}
          breadcrumb="News / About us"
          layout={layout}
          onLayoutChange={setLayout}
          form={<Placeholder screen="Page fields" />}
          preview={<Placeholder screen="Preview" />}
        />
      </div>

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
        <Placeholder screen="Page fields" />
      </PhoneShell>
    </>
  );
}
