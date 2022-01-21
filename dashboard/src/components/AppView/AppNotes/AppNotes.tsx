interface IAppNotesProps {
  title?: string;
  notes?: string | null;
}

function AppNotes(props: IAppNotesProps) {
  const { title, notes } = props;
  return notes ? (
    <>
      <h3 className="section-title">{title ? title : "Installation Notes"}</h3>
      <section className="terminal-wrapper">
        <pre className="terminal-code">{notes}</pre>
      </section>
    </>
  ) : (
    <div />
  );
}

export default AppNotes;
