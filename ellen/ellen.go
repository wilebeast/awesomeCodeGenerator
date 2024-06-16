package ellen

import (
	"context"
	"encoding/json"

	"code.byted.org/gopkg/logs"
)

func Printf(name string, args map[string]interface{}, result map[string]interface{}) {
	argsBytes, _ := json.Marshal(args)
	resultBytes, _ := json.Marshal(result)
	ctxArgs := args["ctx"]
	if ctxArgs != nil {
		if ctx, ok := ctxArgs.(context.Context); ok {
			delete(args, "ctx")
			logs.CtxInfo(ctx, "Calling %s with arguments: %s, result:%s\n", name, string(argsBytes), string(resultBytes))
		}
	} else {
		logs.CtxInfo(context.Background(), "Calling %s with arguments: %s, result:%s\n", name, string(argsBytes), string(resultBytes))
	}

}
