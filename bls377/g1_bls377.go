// Copyright 2020 ConsenSys AG
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

package bls377

// ClearCofactor maps a point in E(Fp) to E(Fp)[r]
// cf https://eprint.iacr.org/2019/403.pdf, 5
func (p *G1Jac) ClearCofactor(a *G1Jac) *G1Jac {
	var res G1Jac
	res.ScalarMultiplication(a, &xGen).Neg(&res).AddAssign(a)
	p.Set(&res)
	return p
}
