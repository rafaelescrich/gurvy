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

// Code generated by gurvy DO NOT EDIT

package bls377

import (
	"fmt"
	"math/big"
	"math/bits"
	"runtime"
	"testing"

	"github.com/consensys/gurvy/bls377/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// utils
func fuzzJacobianG2(p *G2Jac, f *e2) G2Jac {
	var res G2Jac
	res.X.Mul(&p.X, f).Mul(&res.X, f)
	res.Y.Mul(&p.Y, f).Mul(&res.Y, f).Mul(&res.Y, f)
	res.Z.Mul(&p.Z, f)
	return res
}

func fuzzProjectiveG2(p *G2Proj, f *e2) G2Proj {
	var res G2Proj
	res.X.Mul(&p.X, f)
	res.Y.Mul(&p.Y, f)
	res.Z.Mul(&p.Z, f)
	return res
}

func fuzzExtendedJacobianG2(p *g2JacExtended, f *e2) g2JacExtended {
	var res g2JacExtended
	var ff, fff e2
	ff.Square(f)
	fff.Mul(&ff, f)
	res.X.Mul(&p.X, &ff)
	res.Y.Mul(&p.Y, &fff)
	res.ZZ.Mul(&p.ZZ, &ff)
	res.ZZZ.Mul(&p.ZZZ, &fff)
	return res
}

// ------------------------------------------------------------
// tests

func TestG2IsOnCurve(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)
	genFuzz1 := GenE2()
	properties.Property("[BLS377] g2Gen (affine) should be on the curve", prop.ForAll(
		func(a *e2) bool {
			var op1, op2 G2Affine
			op1.FromJacobian(&g2Gen)
			op2.FromJacobian(&g2Gen)
			op2.Y.Mul(&op2.Y, a)
			return op1.IsOnCurve() && !op2.IsOnCurve()
		},
		genFuzz1,
	))

	properties.Property("[BLS377] g2Gen (Jacobian) should be on the curve", prop.ForAll(
		func(a *e2) bool {
			var op1, op2, op3 G2Jac
			op1.Set(&g2Gen)
			op3.Set(&g2Gen)

			op2 = fuzzJacobianG2(&g2Gen, a)
			op3.Y.Mul(&op3.Y, a)
			return op1.IsOnCurve() && op2.IsOnCurve() && !op3.IsOnCurve()
		},
		genFuzz1,
	))

	properties.Property("[BLS377] g2Gen (projective) should be on the curve", prop.ForAll(
		func(a *e2) bool {
			var op1, op2, op3 G2Proj
			op1.FromJacobian(&g2Gen)
			op2.FromJacobian(&g2Gen)
			op3.FromJacobian(&g2Gen)

			op2 = fuzzProjectiveG2(&op1, a)
			op3.Y.Mul(&op3.Y, a)
			return op1.IsOnCurve() && op2.IsOnCurve() && !op3.IsOnCurve()
		},
		genFuzz1,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG2Conversions(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)
	genFuzz1 := GenE2()
	genFuzz2 := GenE2()

	properties.Property("[BLS377] Affine representation should be independent of the Jacobian representative", prop.ForAll(
		func(a *e2) bool {
			g := fuzzJacobianG2(&g2Gen, a)
			var op1 G2Affine
			op1.FromJacobian(&g)
			return op1.X.Equal(&g2Gen.X) && op1.Y.Equal(&g2Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("[BLS377] Affine representation should be independent of a Extended Jacobian representative", prop.ForAll(
		func(a *e2) bool {
			var g g2JacExtended
			g.X.Set(&g2Gen.X)
			g.Y.Set(&g2Gen.Y)
			g.ZZ.Set(&g2Gen.Z)
			g.ZZZ.Set(&g2Gen.Z)
			gfuzz := fuzzExtendedJacobianG2(&g, a)

			var op1 G2Affine
			op1.fromJacExtended(&gfuzz)
			return op1.X.Equal(&g2Gen.X) && op1.Y.Equal(&g2Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("[BLS377] Projective representation should be independent of a Jacobian representative", prop.ForAll(
		func(a *e2) bool {

			g := fuzzJacobianG2(&g2Gen, a)

			var op1 G2Proj
			op1.FromJacobian(&g)
			var u, v e2
			u.Mul(&g.X, &g.Z)
			v.Square(&g.Z).Mul(&v, &g.Z)

			return op1.X.Equal(&u) && op1.Y.Equal(&g.Y) && op1.Z.Equal(&v)
		},
		genFuzz1,
	))

	properties.Property("[BLS377] Jacobian representation should be the same as the affine representative", prop.ForAll(
		func(a *e2) bool {
			var g G2Jac
			var op1 G2Affine
			op1.X.Set(&g2Gen.X)
			op1.Y.Set(&g2Gen.Y)

			var one e2
			one.SetOne()

			g.FromAffine(&op1)

			return g.X.Equal(&g2Gen.X) && g.Y.Equal(&g2Gen.Y) && g.Z.Equal(&one)
		},
		genFuzz1,
	))

	properties.Property("[BLS377] Converting affine symbol for infinity to Jacobian should output correct infinity in Jacobian", prop.ForAll(
		func() bool {
			var g G2Affine
			g.X.SetZero()
			g.Y.SetZero()
			var op1 G2Jac
			op1.FromAffine(&g)
			var one, zero e2
			one.SetOne()
			return op1.X.Equal(&one) && op1.Y.Equal(&one) && op1.Z.Equal(&zero)
		},
	))

	properties.Property("[BLS377] Converting infinity in extended Jacobian to affine should output infinity symbol in Affine", prop.ForAll(
		func() bool {
			var g G2Affine
			var op1 g2JacExtended
			var zero e2
			op1.X.Set(&g2Gen.X)
			op1.Y.Set(&g2Gen.Y)
			g.fromJacExtended(&op1)
			return g.X.Equal(&zero) && g.Y.Equal(&zero)
		},
	))

	properties.Property("[BLS377] Converting infinity in extended Jacobian to Jacobian should output infinity in Jacobian", prop.ForAll(
		func() bool {
			var g G2Jac
			var op1 g2JacExtended
			var zero, one e2
			one.SetOne()
			op1.X.Set(&g2Gen.X)
			op1.Y.Set(&g2Gen.Y)
			g.fromJacExtended(&op1)
			return g.X.Equal(&one) && g.Y.Equal(&one) && g.Z.Equal(&zero)
		},
	))

	properties.Property("[BLS377] [Jacobian] Two representatives of the same class should be equal", prop.ForAll(
		func(a, b *e2) bool {
			op1 := fuzzJacobianG2(&g2Gen, a)
			op2 := fuzzJacobianG2(&g2Gen, b)
			return op1.Equal(&op2)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG2Ops(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)
	genFuzz1 := GenE2()
	genFuzz2 := GenE2()

	genScalar := GenFr()

	properties.Property("[BLS377] [Jacobian] Add should call double when having adding the same point", prop.ForAll(
		func(a, b *e2) bool {
			fop1 := fuzzJacobianG2(&g2Gen, a)
			fop2 := fuzzJacobianG2(&g2Gen, b)
			var op1, op2 G2Jac
			op1.Set(&fop1).AddAssign(&fop2)
			op2.Double(&fop2)
			return op1.Equal(&op2)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.Property("[BLS377] [Jacobian] Adding the opposite of a point to itself should output inf", prop.ForAll(
		func(a, b *e2) bool {
			fop1 := fuzzJacobianG2(&g2Gen, a)
			fop2 := fuzzJacobianG2(&g2Gen, b)
			fop2.Neg(&fop2)
			fop1.AddAssign(&fop2)
			return fop1.Equal(&g2Infinity)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.Property("[BLS377] [Jacobian] Adding the inf to a point should not modify the point", prop.ForAll(
		func(a *e2) bool {
			fop1 := fuzzJacobianG2(&g2Gen, a)
			fop1.AddAssign(&g2Infinity)
			var op2 G2Jac
			op2.Set(&g2Infinity)
			op2.AddAssign(&g2Gen)
			return fop1.Equal(&g2Gen) && op2.Equal(&g2Gen)
		},
		genFuzz1,
	))

	properties.Property("[BLS377] [Jacobian Extended] mAdd (-G) should equal mSub(G)", prop.ForAll(
		func(a *e2) bool {
			fop1 := fuzzJacobianG2(&g2Gen, a)
			var p1, p1Neg G2Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 g2JacExtended
			o1.mAdd(&p1Neg)
			o2.mSub(&p1)

			return o1.X.Equal(&o2.X) &&
				o1.Y.Equal(&o2.Y) &&
				o1.ZZ.Equal(&o2.ZZ) &&
				o1.ZZZ.Equal(&o2.ZZZ)
		},
		genFuzz1,
	))

	properties.Property("[BLS377] [Jacobian Extended] double (-G) should equal doubleNeg(G)", prop.ForAll(
		func(a *e2) bool {
			fop1 := fuzzJacobianG2(&g2Gen, a)
			var p1, p1Neg G2Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 g2JacExtended
			o1.double(&p1Neg)
			o2.doubleNeg(&p1)

			return o1.X.Equal(&o2.X) &&
				o1.Y.Equal(&o2.Y) &&
				o1.ZZ.Equal(&o2.ZZ) &&
				o1.ZZZ.Equal(&o2.ZZZ)
		},
		genFuzz1,
	))

	properties.Property("[BLS377] [Jacobian] Addmix the negation to itself should output 0", prop.ForAll(
		func(a *e2) bool {
			fop1 := fuzzJacobianG2(&g2Gen, a)
			fop1.Neg(&fop1)
			var op2 G2Affine
			op2.FromJacobian(&g2Gen)
			fop1.AddMixed(&op2)
			return fop1.Equal(&g2Infinity)
		},
		genFuzz1,
	))

	properties.Property("[BLS377] scalar multiplication (double and add) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.Modulus()
			var g G2Jac
			g.ScalarMultiplication(&g2Gen, r)

			var scalar, blindedScalard, rminusone big.Int
			var op1, op2, op3, gneg G2Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			op3.ScalarMultiplication(&g2Gen, &rminusone)
			gneg.Neg(&g2Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			op1.ScalarMultiplication(&g2Gen, &scalar)
			op2.ScalarMultiplication(&g2Gen, &blindedScalard)

			return op1.Equal(&op2) && g.Equal(&g2Infinity) && !op1.Equal(&g2Infinity) && gneg.Equal(&op3)

		},
		genScalar,
	))

	properties.Property("[BLS377] psi should map points from E' to itself", prop.ForAll(
		func() bool {
			var a G2Jac
			a.psi(&g2Gen)
			return a.IsOnCurve() && !a.Equal(&g2Gen)
		},
	))

	properties.Property("[BLS377] scalar multiplication (GLV) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.Modulus()
			var g G2Jac
			g.mulGLV(&g2Gen, r)

			var scalar, blindedScalard, rminusone big.Int
			var op1, op2, op3, gneg G2Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			op3.mulGLV(&g2Gen, &rminusone)
			gneg.Neg(&g2Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			op1.mulGLV(&g2Gen, &scalar)
			op2.mulGLV(&g2Gen, &blindedScalard)

			return op1.Equal(&op2) && g.Equal(&g2Infinity) && !op1.Equal(&g2Infinity) && gneg.Equal(&op3)

		},
		genScalar,
	))

	properties.Property("[BLS377] GLV and Double and Add should output the same result", prop.ForAll(
		func(s fr.Element) bool {

			var r big.Int
			var op1, op2 G2Jac
			s.ToBigIntRegular(&r)
			op1.mulWindowed(&g2Gen, &r)
			op2.mulGLV(&g2Gen, &r)
			return op1.Equal(&op2) && !op1.Equal(&g2Infinity)

		},
		genScalar,
	))

	// note : this test is here as we expect to have a different multiExp than the above bucket method
	// for small number of points
	properties.Property("[BLS377] Multi exponentation (<50points) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var g G2Jac
			g.Set(&g2Gen)

			// mixer ensures that all the words of a fpElement are set
			samplePoints := make([]G2Affine, 30)
			sampleScalars := make([]fr.Element, 30)

			for i := 1; i <= 30; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
				samplePoints[i-1].FromJacobian(&g)
				g.AddAssign(&g2Gen)
			}

			var op1MultiExp G2Jac
			op1MultiExp.MultiExp(samplePoints, sampleScalars)

			var finalBigScalar fr.Element
			var finalBigScalarBi big.Int
			var op1ScalarMul G2Jac
			finalBigScalar.SetString("9455").MulAssign(&mixer)
			finalBigScalar.ToBigIntRegular(&finalBigScalarBi)
			op1ScalarMul.ScalarMultiplication(&g2Gen, &finalBigScalarBi)

			return op1ScalarMul.Equal(&op1MultiExp)
		},
		genScalar,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG2MultiExp(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)

	genScalar := GenFr()

	// size of the multiExps
	const nbSamples = 500

	// multi exp points
	var samplePoints [nbSamples]G2Affine
	var g G2Jac
	g.Set(&g2Gen)
	for i := 1; i <= nbSamples; i++ {
		samplePoints[i-1].FromJacobian(&g)
		g.AddAssign(&g2Gen)
	}

	// final scalar to use in double and add method (without mixer factor)
	// n(n+1)(2n+1)/6  (sum of the squares from 1 to n)
	var scalar big.Int
	scalar.SetInt64(nbSamples)
	scalar.Mul(&scalar, new(big.Int).SetInt64(nbSamples+1))
	scalar.Mul(&scalar, new(big.Int).SetInt64(2*nbSamples+1))
	scalar.Div(&scalar, new(big.Int).SetInt64(6))

	properties.Property("[BLS377] Multi exponentation (c=4) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 4)
			result.msmC4(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2Gen, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("[BLS377] Multi exponentation (c=5) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 5)
			result.msmC5(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2Gen, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("[BLS377] Multi exponentation (c=6) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 6)
			result.msmC6(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2Gen, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("[BLS377] Multi exponentation (c=7) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 7)
			result.msmC7(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2Gen, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("[BLS377] Multi exponentation (c=8) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 8)
			result.msmC8(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2Gen, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("[BLS377] Multi exponentation (c=9) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 9)
			result.msmC9(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2Gen, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("[BLS377] Multi exponentation (c=10) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 10)
			result.msmC10(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2Gen, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("[BLS377] Multi exponentation (c=11) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 11)
			result.msmC11(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2Gen, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("[BLS377] Multi exponentation (c=12) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 12)
			result.msmC12(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2Gen, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("[BLS377] Multi exponentation (c=13) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 13)
			result.msmC13(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2Gen, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("[BLS377] Multi exponentation (c=14) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 14)
			result.msmC14(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2Gen, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("[BLS377] Multi exponentation (c=15) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 15)
			result.msmC15(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2Gen, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	if !testing.Short() {

		properties.Property("[BLS377] Multi exponentation (c=16) should be consistant with sum of square", prop.ForAll(
			func(mixer fr.Element) bool {

				var result, expected G2Jac

				// mixer ensures that all the words of a fpElement are set
				var sampleScalars [nbSamples]fr.Element

				for i := 1; i <= nbSamples; i++ {
					sampleScalars[i-1].SetUint64(uint64(i)).
						MulAssign(&mixer).
						FromMont()
				}

				// semaphore to limit number of cpus
				opt := NewMultiExpOptions(runtime.NumCPU())
				opt.lock.Lock()
				scalars := partitionScalars(sampleScalars[:], 16)
				result.msmC16(samplePoints[:], scalars, opt)

				// compute expected result with double and add
				var finalScalar, mixerBigInt big.Int
				finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
				expected.ScalarMultiplication(&g2Gen, &finalScalar)

				return result.Equal(&expected)
			},
			genScalar,
		))

	}

	if !testing.Short() {

		properties.Property("[BLS377] Multi exponentation (c=20) should be consistant with sum of square", prop.ForAll(
			func(mixer fr.Element) bool {

				var result, expected G2Jac

				// mixer ensures that all the words of a fpElement are set
				var sampleScalars [nbSamples]fr.Element

				for i := 1; i <= nbSamples; i++ {
					sampleScalars[i-1].SetUint64(uint64(i)).
						MulAssign(&mixer).
						FromMont()
				}

				// semaphore to limit number of cpus
				opt := NewMultiExpOptions(runtime.NumCPU())
				opt.lock.Lock()
				scalars := partitionScalars(sampleScalars[:], 20)
				result.msmC20(samplePoints[:], scalars, opt)

				// compute expected result with double and add
				var finalScalar, mixerBigInt big.Int
				finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
				expected.ScalarMultiplication(&g2Gen, &finalScalar)

				return result.Equal(&expected)
			},
			genScalar,
		))

	}

	if !testing.Short() {

		properties.Property("[BLS377] Multi exponentation (c=21) should be consistant with sum of square", prop.ForAll(
			func(mixer fr.Element) bool {

				var result, expected G2Jac

				// mixer ensures that all the words of a fpElement are set
				var sampleScalars [nbSamples]fr.Element

				for i := 1; i <= nbSamples; i++ {
					sampleScalars[i-1].SetUint64(uint64(i)).
						MulAssign(&mixer).
						FromMont()
				}

				// semaphore to limit number of cpus
				opt := NewMultiExpOptions(runtime.NumCPU())
				opt.lock.Lock()
				scalars := partitionScalars(sampleScalars[:], 21)
				result.msmC21(samplePoints[:], scalars, opt)

				// compute expected result with double and add
				var finalScalar, mixerBigInt big.Int
				finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
				expected.ScalarMultiplication(&g2Gen, &finalScalar)

				return result.Equal(&expected)
			},
			genScalar,
		))

	}

	if !testing.Short() {

		properties.Property("[BLS377] Multi exponentation (c=22) should be consistant with sum of square", prop.ForAll(
			func(mixer fr.Element) bool {

				var result, expected G2Jac

				// mixer ensures that all the words of a fpElement are set
				var sampleScalars [nbSamples]fr.Element

				for i := 1; i <= nbSamples; i++ {
					sampleScalars[i-1].SetUint64(uint64(i)).
						MulAssign(&mixer).
						FromMont()
				}

				// semaphore to limit number of cpus
				opt := NewMultiExpOptions(runtime.NumCPU())
				opt.lock.Lock()
				scalars := partitionScalars(sampleScalars[:], 22)
				result.msmC22(samplePoints[:], scalars, opt)

				// compute expected result with double and add
				var finalScalar, mixerBigInt big.Int
				finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
				expected.ScalarMultiplication(&g2Gen, &finalScalar)

				return result.Equal(&expected)
			},
			genScalar,
		))

	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG2CofactorCleaning(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS377] Clearing the cofactor of a random point should set it in the r-torsion", prop.ForAll(
		func() bool {
			var a, x, b e2
			a.SetRandom()

			x.Square(&a).Mul(&x, &a).Add(&x, &bTwistCurveCoeff)
			for x.Legendre() != 1 {
				a.SetRandom()
				x.Square(&a).Mul(&x, &a).Add(&x, &bTwistCurveCoeff)
			}

			b.Sqrt(&x)
			var point, pointCleared, infinity G2Jac
			point.X.Set(&a)
			point.Y.Set(&b)
			point.Z.SetOne()
			pointCleared.ClearCofactor(&point)
			infinity.Set(&g2Infinity)
			return point.IsOnCurve() && pointCleared.IsInSubGroup() && !pointCleared.Equal(&infinity)
		},
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))

}

func TestG2BatchScalarMultiplication(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)

	genScalar := GenFr()

	// size of the multiExps
	const nbSamples = 500

	properties.Property("[BLS377] BatchScalarMultiplication should be consistant with individual scalar multiplications", prop.ForAll(
		func(mixer fr.Element) bool {
			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			result := BatchScalarMultiplicationG2(&g2GenAff, sampleScalars[:])

			if len(result) != len(sampleScalars) {
				return false
			}

			for i := 0; i < len(result); i++ {
				var expectedJac G2Jac
				var expected G2Affine
				var b big.Int
				expectedJac.mulGLV(&g2Gen, sampleScalars[i].ToBigInt(&b))
				expected.FromJacobian(&expectedJac)
				if !result[i].Equal(&expected) {
					return false
				}
			}
			return true
		},
		genScalar,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkG2BatchScalarMul(b *testing.B) {
	// ensure every words of the scalars are filled
	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	const pow = 15
	const nbSamples = 1 << pow

	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
	}

	for i := 5; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_ = BatchScalarMultiplicationG2(&g2GenAff, sampleScalars[:using])
			}
		})
	}
}

func BenchmarkG2ScalarMul(b *testing.B) {

	var scalar big.Int
	r := fr.Modulus()
	scalar.SetString("5243587517512619047944770508185965837690552500527637822603658699938581184513", 10)
	scalar.Add(&scalar, r)

	var doubleAndAdd G2Jac

	b.Run("double and add", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			doubleAndAdd.ScalarMultiplication(&g2Gen, &scalar)
		}
	})

	var glv G2Jac
	b.Run("GLV", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			glv.mulGLV(&g2Gen, &scalar)
		}
	})

}

func BenchmarkG2CofactorClearing(b *testing.B) {
	var a G2Jac
	a.Set(&g2Gen)
	for i := 0; i < b.N; i++ {
		a.ClearCofactor(&a)
	}
}

func BenchmarkG2Add(b *testing.B) {
	var a G2Jac
	a.Double(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddAssign(&g2Gen)
	}
}

func BenchmarkG2mAdd(b *testing.B) {
	var a g2JacExtended
	a.double(&g2GenAff)

	var c G2Affine
	c.FromJacobian(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.mAdd(&c)
	}

}

func BenchmarkG2AddMixed(b *testing.B) {
	var a G2Jac
	a.Double(&g2Gen)

	var c G2Affine
	c.FromJacobian(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddMixed(&c)
	}

}

func BenchmarkG2Double(b *testing.B) {
	var a G2Jac
	a.Set(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.DoubleAssign()
	}

}

func BenchmarkG2MultiExpG2(b *testing.B) {
	// ensure every words of the scalars are filled
	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	const pow = (bits.UintSize / 2) - (bits.UintSize / 8) // 24 on 64 bits arch, 12 on 32 bits
	const nbSamples = 1 << pow

	var samplePoints [nbSamples]G2Affine
	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
		samplePoints[i-1] = g2GenAff
	}

	var testPoint G2Jac

	for i := 5; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				testPoint.MultiExp(samplePoints[:using], sampleScalars[:using])
			}
		})
	}
}
