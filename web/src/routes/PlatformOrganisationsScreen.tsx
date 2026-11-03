import { useState } from "react";

import { NewOrganisation } from "./AdministrationPage";
import { PlatformOrganisations } from "./PlatformOrganisations";

/**
 * The organisations, with the form that starts one. The form is a dialog
 * rather than a card on the screen, because starting an organisation is
 * something a platform admin does now and then, and the list is what they
 * came for.
 *
 * chatcms-audience: implementer
 */
export function PlatformOrganisationsScreen() {
  const [making, setMaking] = useState(false);
  const [made, setMade] = useState(0);

  return (
    <div className="flex flex-col gap-6">
      {making && (
        <NewOrganisation
          onMade={() => {
            setMaking(false);
            setMade((count) => count + 1);
          }}
        />
      )}

      <PlatformOrganisations key={made} onNew={() => setMaking(true)} />
    </div>
  );
}
