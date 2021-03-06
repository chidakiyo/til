import "tslib";
import "skatejs-web-components";
import * as skate from "skatejs";

(window as any).__extends = function(d: any, b: any) {
    Object.setPrototypeOf(d, b);
    var __: any = function() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};

import "./prop-examples";

class CountUpComponent extends skate.Component implements skate.OnRenderCallback {
    static get props(): { [key: string]: skate.PropAttr<CountUpComponent, any>; } {
        return {
            count: skate.prop.number({
                attribute: true,
                default(elem, data) {
                    return 7;
                },
            }),
        }
    }

    count: number;

    click() {
        this.count += 1;
    }

    renderCallback(): any {
        return (
            <div>
                <p>Count: <span>{this.count}</span></p>
                <button onClick={e => this.click()}>Count up</button>
            </div>
        );
    }
}
customElements.define("x-countup", CountUpComponent);

customElements.define("x-app", class extends skate.Component implements skate.OnRenderCallback {
    renderCallback() {
        // for https://github.com/Microsoft/TypeScript/issues/7004
        // 型チェックが行われると大量にエラーが出るのでanyで殺す 堅牢さに影響はないはず
        const anyProps: any = {};
        return (
            <div>
                <h1>app</h1>
                <CountUpComponent count={100} {...anyProps}></CountUpComponent>
            </div>
        );
    }
});
