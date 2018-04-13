const initialState: string = localStorage.getItem("kubeapps_namespace") || "default";

const namespaceReducer = (state: string = initialState): string => {
  return state;
};

export default namespaceReducer;
