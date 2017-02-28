// Copyright © 2016 Abcum Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"fmt"
	"github.com/abcum/surreal/sql"
	"github.com/abcum/surreal/util/item"
	"github.com/abcum/surreal/util/keys"
)

func (e *executor) executeDeleteStatement(ast *sql.DeleteStatement) (out []interface{}, err error) {

	for k, w := range ast.What {
		if what, ok := w.(*sql.Param); ok {
			ast.What[k] = e.get(what.ID)
		}
	}

	for _, w := range ast.What {

		switch what := w.(type) {

		default:
			return out, fmt.Errorf("Can not execute DELETE query using value '%v' with type '%T'", what, what)

		case *sql.Thing:
			key := &keys.Thing{KV: ast.KV, NS: ast.NS, DB: ast.DB, TB: what.TB, ID: what.ID}
			kv, _ := e.txn.Get(0, key.Encode())
			doc := item.New(kv, e.txn, key, e.ctx)
			if ret, err := delete(doc, ast); err != nil {
				return nil, err
			} else if ret != nil {
				out = append(out, ret)
			}

		case *sql.Table:
			key := &keys.Table{KV: ast.KV, NS: ast.NS, DB: ast.DB, TB: what.TB}
			kvs, _ := e.txn.GetL(0, key.Encode())
			for _, kv := range kvs {
				doc := item.New(kv, e.txn, nil, e.ctx)
				if ret, err := delete(doc, ast); err != nil {
					return nil, err
				} else if ret != nil {
					out = append(out, ret)
				}
			}

		}

	}

	return

}

func delete(doc *item.Doc, ast *sql.DeleteStatement) (out interface{}, err error) {

	if !doc.Check(ast.Cond) {
		return
	}

	if !doc.Allow("DELETE") {
		return
	}

	if err = doc.Erase(); err != nil {
		return
	}

	if err = doc.PurgeIndex(); err != nil {
		return
	}

	if err = doc.PurgeThing(); err != nil {
		return
	}

	if ast.Hard {

		if err = doc.PurgePatch(); err != nil {
			return
		}

	}

	out = doc.Yield(ast.Echo)

	return

}
