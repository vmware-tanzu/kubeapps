import { RouterAction } from "connected-react-router";
import * as React from "react";

import LoadingWrapper from "../../components/LoadingWrapper";
import PageHeader from "../../components/PageHeader";
import { IServiceCatalogState } from "../../reducers/catalog";
import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IRBACRole } from "../../shared/types";
import { ErrorSelector } from "../ErrorAlert";
import ServiceClassPlan from "./ServiceClassPlan";

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

interface IServiceClassViewProps {
  classes: IServiceCatalogState["classes"];
  classname: string;
  createError?: Error;
  error?: Error;
  getClasses: () => Promise<any>;
  getPlans: () => Promise<any>;
  plans: IServiceCatalogState["plans"];
  provision: (
    instanceName: string,
    namespace: string,
    className: string,
    planName: string,
    parameters: {},
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
  namespace: string;
}

class ServiceClassView extends React.Component<IServiceClassViewProps> {
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
      namespace,
    } = this.props;
    const loaded = !classes.isFetching && !plans.isFetching;
    let classPlans: JSX.Element[] = [];
    let svcClass: IClusterServiceClass | undefined;

    if (loaded) {
      svcClass =
        classes.list.find(potential => !!(potential.spec.externalName === classname)) || undefined;
      if (svcClass) {
        const foundSVCClass = svcClass;
        const filteredPlans = plans.list.filter(
          plan => plan.spec.clusterServiceClassRef.name === foundSVCClass.metadata.name,
        );
        classPlans = filteredPlans.map(plan => {
          return (
            <ServiceClassPlan
              key={plan.metadata.name}
              svcClass={foundSVCClass}
              svcPlan={plan}
              provision={provision}
              push={push}
              createError={createError}
              namespace={namespace}
            />
          );
        });
      }
    }

    return (
      <div className="container">
        <PageHeader>
          <h1>Plans: {classname}</h1>
        </PageHeader>
        <main>
          <LoadingWrapper loaded={loaded}>
            <div className="class-view">
              {error && (
                <ErrorSelector
                  error={error}
                  resource="Service Plans"
                  action="list"
                  defaultRequiredRBACRoles={{ list: RequiredRBACRoles }}
                />
              )}
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
                  <tbody>{classPlans}</tbody>
                </table>
              </div>
            </div>
          </LoadingWrapper>
        </main>
      </div>
    );
  }
}

export default ServiceClassView;
