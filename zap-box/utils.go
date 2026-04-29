// Copyright 2019 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zapbox

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// WithLevel returns a new context derived from ctx
// that has a logger that only logs messages at or above
// the given level.
//
// *Important!* This does not work with all underlying core
//
//	implementation. See https://github.com/uber-go/zap/issues/581#issuecomment-600641485
//	for details.
func WithLevel(level zapcore.Level) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &coreWithLevel{
			Core:  core,
			level: level,
		}
	})
}

type coreWithLevel struct {
	zapcore.Core
	level zapcore.Level
}

func (c *coreWithLevel) Enabled(level zapcore.Level) bool {
	return c.level.Enabled(level)
}

func (c *coreWithLevel) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.level.Enabled(e.Level) {
		return ce
	}

	return ce.AddCore(e, c.Core)
}

func (c *coreWithLevel) With(fields []zap.Field) zapcore.Core {
	return &coreWithLevel{
		Core:  c.Core.With(fields),
		level: c.level,
	}
}
