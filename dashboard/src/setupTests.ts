import "raf/polyfill"; // polyfill for requestAnimationFrame

import { configure } from "enzyme";
import * as Adapter from "enzyme-adapter-react-16";

configure({ adapter: new Adapter() });
