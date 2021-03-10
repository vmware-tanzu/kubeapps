import PageHeader from "components/PageHeader/PageHeader";
import placeholder from "../../placeholder.png";

interface IOperatorHeaderProps {
  title: string;
  version?: string;
  icon?: string;
  buttons?: JSX.Element[];
}

export default function OperatorHeader(props: IOperatorHeaderProps) {
  const { title, icon, version, buttons } = props;
  return (
    <PageHeader
      title={title}
      titleSize="md"
      icon={icon || placeholder}
      operator={true}
      version={
        version ? (
          <div className="header-version">
            <label className="header-version-label">Operator Version: {version}</label>
          </div>
        ) : undefined
      }
      buttons={buttons}
    />
  );
}
