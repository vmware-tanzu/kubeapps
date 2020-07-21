import * as React from "react";

import "./AppNotes.v2.css";

interface IAppNotesProps {
  title?: string;
  notes?: string | null;
}

function AppNotes(props: IAppNotesProps) {
  const { title, notes } = props;
  return notes ? (
    <>
      <h5 className="section-title">{title ? title : "Installation Notes"}</h5>
      <section className="app-notes">
        <pre className="terminal-code">{notes}</pre>
      </section>
    </>
  ) : (
    <div />
  );
}

export default AppNotes;
