import { Navigate, Route, Routes } from "react-router";

import { TooltipProvider } from "./components/ui/tooltip";
import { Toaster } from "./components/ui/toast";
import { EditorPage } from "./routes/EditorPage";
import { NotFound } from "./routes/NotFound";
import { OrganisationPage } from "./routes/OrganisationPage";
import { PeoplePage } from "./routes/PeoplePage";
import { ProjectPage } from "./routes/ProjectPage";
import { ProjectsPage } from "./routes/ProjectsPage";
import { SignInPage } from "./routes/SignInPage";

export function App() {
  return (
    <TooltipProvider delayDuration={200}>
      <Routes>
        <Route path="/" element={<Navigate to="/projects" replace />} />
        <Route path="/sign-in" element={<SignInPage />} />
        <Route path="/projects" element={<ProjectsPage />} />
        <Route path="/o/:organisation/projects" element={<ProjectsPage />} />
        <Route path="/help" element={<OrganisationPage title="Help" />} />

        <Route path="/p/:project/setup" element={<ProjectPage title="Set up a project" />} />
        <Route
          path="/p/:project/people"
          element={<ProjectPage title="People & roles" body={<PeoplePage organisation="technative" />} />}
        />
        <Route path="/p/:project/settings" element={<ProjectPage title="Project settings" />} />
        <Route path="/p/:project/workflows" element={<ProjectPage title="Workflows" />} />
        <Route path="/p/:project/content" element={<ProjectPage title="Content" />} />
        <Route path="/p/:project/content/:page" element={<EditorPage />} />
        <Route path="/p/:project/media" element={<ProjectPage title="Media" />} />
        <Route path="/p/:project/publish" element={<ProjectPage title="Publish" />} />
        <Route path="/p/:project/history" element={<ProjectPage title="History" />} />
        <Route path="/p/:project/sync" element={<ProjectPage title="Sync status" />} />

        <Route path="*" element={<NotFound />} />
      </Routes>
      <Toaster />
    </TooltipProvider>
  );
}
