import * as React from "react";

interface IAppNotesProps {
  notes?: string | null;
}

class AppNotes extends React.PureComponent<IAppNotesProps> {
  public render() {
    const { notes } = this.props;
    return notes ? (
      <React.Fragment>
        <h6>Notes</h6>
        <section className="AppNotes Terminal elevation-1">
          <div className="Terminal__Top type-small">
            <div className="Terminal__Top__Buttons">
              <span className="Terminal__Top__Button Terminal__Top__Button--red" />
              <span className="Terminal__Top__Button Terminal__Top__Button--yellow" />
              <span className="Terminal__Top__Button Terminal__Top__Button--green" />
            </div>
          </div>
          <div className="Terminal__Tab">
            <pre className="Terminal__Code">
              <code>{notes}</code>
            </pre>
          </div>
        </section>
      </React.Fragment>
    ) : (
      ""
    );
  }
}

export default AppNotes;
