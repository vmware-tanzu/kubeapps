import { RouterAction } from "connected-react-router";
import * as React from "react";
import { IServiceCatalogState } from "reducers/catalog";
import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { definedNamespaces } from "../../shared/Namespace";
import { IServicePlan } from "../../shared/ServiceCatalog";
import { ForbiddenError, IRBACRole } from "../../shared/types";
import { PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";
import ProvisionButton from "./ProvisionButton";

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "servicecatalog.k8s.io",
    clusterWide: true,
    resource: "clusterserviceclasses",
    verbs: ["list"],
  },
  {
    apiGroup: "servicecatalog.k8s.io",
    clusterWide: true,
    resource: "clusterserviceplans",
    verbs: ["list"],
  },
];

interface IClassViewProps {
  classes: IServiceCatalogState["classes"];
  classname: string;
  createError: Error;
  error: Error;
  getClasses: () => Promise<any>;
  getPlans: () => Promise<any>;
  plans: IServicePlan[];
  provision: (
    instanceName: string,
    namespace: string,
    className: string,
    planName: string,
    parameters: {},
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
  svcClass: IClusterServiceClass | undefined;
  namespace: string;
}

class ClassView extends React.Component<IClassViewProps> {
  public componentDidMount() {
    this.props.getClasses();
    this.props.getPlans();
  }

  public render() {
    const {
      createError,
      classes,
      classname,
      error,
      plans,
      provision,
      push,
      svcClass,
      namespace,
    } = this.props;
    const classPlans = svcClass
      ? plans.filter(plan => plan.spec.clusterServiceClassRef.name === svcClass.metadata.name)
      : [];

    return (
      <div className="class-view">
        <h1>Plans: {classname}</h1>
        {error ? (
          this.renderError()
        ) : (
          <div>
            <p>Service Plans available for provisioning under {classname}</p>
            <table className="striped">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Specs</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {svcClass &&
                  classPlans.map(plan => {
                    // TODO: Check classes.isFetching
                    const serviceClass = classes.list.find(
                      potential =>
                        potential.metadata.name === plan.spec.clusterServiceClassRef.name,
                    );
                    const { spec } = plan;
                    const { externalMetadata } = spec;
                    const name = externalMetadata
                      ? externalMetadata.displayName
                      : spec.externalName;
                    const description =
                      externalMetadata && externalMetadata.bullets
                        ? externalMetadata.bullets
                        : [spec.description];
                    const bullets = (
                      <div>
                        <ul className="margin-reset">
                          {description.map(bullet => (
                            <li key={bullet}>{bullet}</li>
                          ))}
                        </ul>
                      </div>
                    );

                    return (
                      <tr key={plan.spec.externalID}>
                        <td>
                          <b>{name}</b>
                        </td>
                        <td>{bullets}</td>
                        <td className="text-r">
                          <ProvisionButton
                            selectedClass={serviceClass}
                            selectedPlan={plan}
                            provision={provision}
                            push={push}
                            namespace={namespace}
                            error={createError}
                          />
                        </td>
                      </tr>
                    );
                  })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    );
  }

  // TODO: Replace with ErrorSelector
  private renderError() {
    const { error } = this.props;
    return error instanceof ForbiddenError ? (
      <PermissionsErrorAlert
        action="list Service Plans"
        roles={RequiredRBACRoles}
        namespace={definedNamespaces.all}
      />
    ) : (
      <UnexpectedErrorAlert />
    );
  }
}

export default ClassView;
