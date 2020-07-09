import * as React from "react";

import { ILoginFormProps } from "./LoginForm.v2";

const LoginForm = React.lazy(() => import("./LoginForm"));
const LoginFormV2 = React.lazy(() => import("./LoginForm.v2"));

interface ILoginFormSelectorProps extends ILoginFormProps {
  UI: string;
}

const LoginFormSelector: React.FC<ILoginFormSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <LoginFormV2 {...props} /> : <LoginForm {...props} />}
  </React.Suspense>
);

export default LoginFormSelector;
