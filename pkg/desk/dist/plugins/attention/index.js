// node_modules/svelte/src/constants.js
var e = {}, t = Symbol("uninitialized"), n = "http://www.w3.org/1999/xhtml", r = "http://www.w3.org/2000/svg", i = "http://www.w3.org/1998/Math/MathML", a = Array.isArray, o = Array.prototype.indexOf, s = Array.prototype.includes, c = Array.from, l = Object.defineProperty, u = Object.getOwnPropertyDescriptor, d = Object.getOwnPropertyDescriptors, f = Object.prototype, p = Array.prototype, m = Object.getPrototypeOf, h = Object.isExtensible, g = () => {};
function _(e) {
	for (var t = 0; t < e.length; t++) e[t]();
}
function v() {
	var e, t;
	return {
		promise: new Promise((n, r) => {
			e = n, t = r;
		}),
		resolve: e,
		reject: t
	};
}
var y = 1024, b = 2048, x = 4096, S = 8192, C = 16384, ee = 32768, te = 1 << 25, ne = 65536, w = 1 << 19, re = 1 << 20, ie = 1 << 25, ae = 1 << 21, oe = 1 << 22, se = 1 << 23, ce = Symbol("$state"), le = Symbol("component"), ue = Symbol("legacy props"), de = Symbol(""), fe = Symbol("attributes"), pe = Symbol("class"), me = Symbol("style"), he = Symbol("text"), ge = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), T = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function _e() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function ve(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function ye() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var E = !1;
function be(e) {
	E = e;
}
var D;
function xe(t) {
	if (t === null) throw ve(), e;
	return D = t;
}
function Se() {
	return xe(/* @__PURE__ */ Wt(D));
}
function O(t) {
	if (E) {
		if (/* @__PURE__ */ Wt(D) !== null) throw ve(), e;
		D = t;
	}
}
function k(e = 1) {
	if (E) {
		for (var t = e, n = D; t--;) n = /* @__PURE__ */ Wt(n);
		D = n;
	}
}
function Ce(e = !0) {
	for (var t = 0, n = D;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ Wt(n);
		e && n.remove(), n = i;
	}
}
function A(t) {
	if (!t || t.nodeType !== 8) throw ve(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function we(e) {
	return e === this.v;
}
function Te(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function Ee(e) {
	return !Te(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function De() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Oe(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function ke(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function Ae() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function je(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Me() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Ne(e) {
	throw Error("https://svelte.dev/e/props_invalid_value");
}
function Pe() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Fe() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Ie() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Le() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var Re = null;
function ze(e) {
	Re = e;
}
function Be(e, t = !1, n) {
	Re = {
		p: Re,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: U,
		l: null
	};
}
function Ve(e) {
	var t = Re, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) an(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, Re = t.p, He(e);
}
function He(e = {}) {
	return l(e, le, { value: !0 }), e;
}
function Ue() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var We = [];
function Ge() {
	var e = We;
	We = [], _(e);
}
function Ke(e) {
	if (We.length === 0 && !mt) {
		var t = We;
		queueMicrotask(() => {
			t === We && Ge();
		});
	}
	We.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var qe = ~(b | x | y);
function j(e, t) {
	e.f = e.f & qe | t;
}
function Je(e) {
	e.f & 512 || e.deps === null ? j(e, y) : j(e, x);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Ye(e, t) {
	e.f & 2048 ? t.set(e, b) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, x), j(e, y);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Xe(e) {
	var t = H, n = U;
	En(null), Dn(null);
	try {
		return e();
	} finally {
		En(t), Dn(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function Ze(e, t, n, r) {
	let i = Ue() ? tt : at;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = U, c = Qe(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				Zt(e, s);
			}
			$e();
		}
	}
	var d = et();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ rt(e))).then(u).catch((e) => Zt(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), $e());
	}) : f();
}
function Qe() {
	var e = U, t = H, n = Re, r = M;
	return function(i = !0) {
		Dn(e), En(t), ze(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function $e(e = !0) {
	Dn(null), En(null), ze(null), e && M?.deactivate();
}
function et() {
	var e = U, t = e.b, n = M, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function tt(e) {
	var n = 2 | b;
	return U !== null && (U.f |= w), {
		ctx: Re,
		deps: null,
		effects: null,
		equals: we,
		f: n,
		fn: e,
		reactions: null,
		rv: 0,
		v: t,
		wv: 0,
		parent: U,
		ac: null
	};
}
var nt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function rt(e, n, r) {
	let i = U;
	i === null && De();
	var a = void 0, o = At(t), s = !H, c = /* @__PURE__ */ new Set();
	return sn(() => {
		var t = U, n = v();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== ge && n.reject(e);
			}).finally($e);
		} catch (e) {
			n.reject(e), $e();
		}
		var r = M;
		if (s) {
			if (t.f & 32768) var l = et();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(nt);
			else for (let e of c.values()) e.reject(nt);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== nt && (r.activate(), t ? (o.f |= se, Pt(o, t)) : (o.f & 8388608 && (o.f ^= se), Pt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), nn(() => {
		for (let e of c) e.reject(nt);
	}), new Promise((e) => {
		function t(n) {
			function r() {
				n === a ? e(o) : t(a);
			}
			n.then(r, r);
		}
		t(a);
	});
}
/*#__NO_SIDE_EFFECTS__*/
function it(e) {
	let t = /* @__PURE__ */ tt(e);
	return kn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function at(e) {
	let t = /* @__PURE__ */ tt(e);
	return t.equals = Ee, t;
}
function ot(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) mn(t[n]);
	}
}
function st(e) {
	var n, r = U, i = e.parent;
	if (!Cn && i !== null && e.v !== t && i.f & 24576) return _e(), e.v;
	Dn(i);
	try {
		ot(e), n = Vn(e);
	} finally {
		Dn(r);
	}
	return n;
}
function ct(e) {
	var t = st(e);
	!e.equals(t) && (e.wv = Rn(), (!M?.is_fork || e.deps === null) && (M === null ? e.v = t : (M.capture(e, t, !0), ft?.capture(e, t, !0)), e.deps === null)) ? j(e, y) : Cn || (pt === null ? Je(e) : (tn() || M?.is_fork) && pt.set(e, t));
}
function lt(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Xe(() => {
		t.ac.abort(ge), t.ac = null;
	}), t.fn !== null && (t.teardown = g), Wn(t, 0), fn(t));
}
function ut(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Gn(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var dt = null, M = null, ft = null, pt = null, N = null, mt = !1, ht = !1, gt = null, _t = null, vt = 0, yt = 1, bt = class e {
	id = yt++;
	started = !1;
	linked = !0;
	prev = null;
	next = null;
	async_deriveds = /* @__PURE__ */ new Map();
	current = /* @__PURE__ */ new Map();
	previous = /* @__PURE__ */ new Map();
	#e = /* @__PURE__ */ new Set();
	#t = /* @__PURE__ */ new Set();
	#n = 0;
	#r = /* @__PURE__ */ new Map();
	#i = null;
	#a = [];
	#o = [];
	#s = /* @__PURE__ */ new Map();
	#c = /* @__PURE__ */ new Map();
	#l = /* @__PURE__ */ new Set();
	is_fork = !1;
	#u = !1;
	constructor() {
		dt === null ? dt = this : (dt.next = this, this.prev = dt), dt = this;
	}
	#d() {
		if (this.is_fork) return !0;
		for (let e of this.#r.keys()) if (!this.#f(e)) return !0;
		return !1;
	}
	#f(e) {
		for (var t = e; t !== null; t = t.parent) if (this.#c.has(t)) return !0;
		return !1;
	}
	skip_effect(e) {
		this.#c.has(e) || this.#c.set(e, {
			d: [],
			m: []
		}), this.#l.delete(e);
	}
	unskip_effect(e, t = (e) => this.schedule(e)) {
		var n = this.#c.get(e);
		if (n) {
			this.#c.delete(e);
			for (var r of n.d) j(r, b), t(r);
			for (r of n.m) j(r, x), t(r);
		}
		this.#l.add(e);
	}
	#p() {
		var e = [];
		for (let i of this.#a) if (!(i.f & 16384 || !(i.f & 6144))) {
			for (var t = i, n = !1; t.parent !== null;) {
				t = t.parent;
				var r = t.f;
				if (r & 96) {
					if (!(r & 1024)) {
						n = !0;
						break;
					}
					t.f ^= y;
				}
			}
			n || e.push(t);
		}
		return this.#a = [], e;
	}
	#m() {
		this.started = !0;
		for (let [e, t] of this.#s) j(e, t), this.schedule(e);
		this.apply();
		for (var t = gt = [], n = [], r = _t = []; this.#a.length > 0;) {
			vt++ > 1e3 && (this.#b(), xt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Et(e), this.#d() || this.discard(), t;
			}
		}
		if (M = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (gt = null, _t = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Tt(e, t);
			r.length > 0 && M.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), ft = this, Ct(n), Ct(t), ft = null, this.#i?.resolve();
			var o = M;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Ot.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= y;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= y : i & 4 ? t.push(r) : zn(r) && (i & 16 && this.add_dirty_reaction(r, x), Gn(r));
				var o = r.first;
				if (o !== null) {
					r = o;
					continue;
				}
			}
			for (; r !== null;) {
				var s = r.next;
				if (s !== null) {
					r = s;
					break;
				}
				r = r.parent;
			}
		}
	}
	#g() {
		for (var e = this.prev; e !== null; e = e.prev) if (!e.is_fork && this.#_(e)) return e;
		return null;
	}
	#_(e) {
		for (let [t, [, n]] of this.current) if (!n && e.current.has(t)) return !0;
		return !1;
	}
	#v(e) {
		for (let [t, n] of e.current) !this.previous.has(t) && e.previous.has(t) && this.previous.set(t, e.previous.get(t)), this.current.set(t, n);
		for (let [t, n] of e.async_deriveds) {
			let e = this.async_deriveds.get(t);
			e && n.promise.then(e.resolve).catch(e.reject);
		}
		e.async_deriveds.clear(), this.transfer_reactions(e.#s);
		let t = (e) => {
			var n = e.reactions;
			if (n !== null && !(e.f & 2 && !(e.f & 6144))) for (let e of n) {
				var r = e.f;
				if (r & 2) t(e);
				else {
					var i = e;
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), j(i, b), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), M = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) Ye(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), pt?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		M = this;
	}
	deactivate() {
		M = null, pt = null;
	}
	flush() {
		try {
			ht = !0, M = this, this.#m();
		} finally {
			vt = 0, N = null, gt = null, _t = null, ht = !1, M = null, pt = null, Ot.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(nt);
		this.#b(), this.#i?.resolve();
	}
	register_created_effect(e) {
		this.#o.push(e);
	}
	increment(e, t) {
		if (this.#n += 1, e) {
			let e = this.#r.get(t) ?? 0;
			this.#r.set(t, e + 1);
		}
	}
	decrement(e, t) {
		if (--this.#n, e) {
			let e = this.#r.get(t) ?? 0;
			e === 1 ? this.#r.delete(t) : this.#r.set(t, e - 1);
		}
		this.#u || (this.#u = !0, Ke(() => {
			this.#u = !1, this.linked && this.flush();
		}));
	}
	transfer_reactions(e) {
		for (let [t, n] of e) this.add_dirty_reaction(t, n);
		e.clear();
	}
	add_dirty_reaction(e, t) {
		(t === 2048 || this.#s.get(e) !== 2048) && this.#s.set(e, t);
	}
	oncommit(e) {
		this.#e.add(e);
	}
	ondiscard(e) {
		this.#t.add(e);
	}
	settled() {
		return (this.#i ??= v()).promise;
	}
	static ensure() {
		if (M === null) {
			let t = M = new e();
			!ht && Ke(() => {
				t.started || t.flush();
			});
		}
		return M;
	}
	apply() {
		pt = null;
	}
	schedule(e) {
		N = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? dt = e : t.prev = e, this.linked = !1;
		}
	}
};
function xt() {
	try {
		Me();
	} catch (e) {
		Zt(e, N);
	}
}
var St = null;
function Ct(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && zn(r) && (St = /* @__PURE__ */ new Set(), Gn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && gn(r), St?.size > 0)) {
				Ot.clear();
				for (let e of St) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) St.has(n) && (St.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Gn(n);
					}
				}
				St.clear();
			}
		}
		St = null;
	}
}
function wt(e) {
	M.schedule(e);
}
function Tt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), j(e, y);
		for (var n = e.first; n !== null;) Tt(n, t), n = n.next;
	}
}
function Et(e) {
	j(e, y);
	for (var t = e.first; t !== null;) Et(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Dt = /* @__PURE__ */ new Set(), Ot = /* @__PURE__ */ new Map(), kt = !1;
function At(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: we,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function P(e, t) {
	let n = At(e, t);
	return kn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function jt(e, t = !1, n = !0) {
	let r = At(e);
	return t || (r.equals = Ee), r;
}
function F(e, t, n = !1) {
	return H !== null && (!Tn || H.f & 131072) && Ue() && H.f & 4325394 && (On === null || !On.has(e)) && Ie(), Pt(e, n ? Rt(t) : t, _t);
}
var Mt = null, Nt = 0;
function Pt(e, t, n = null) {
	if (!e.equals(t)) {
		Cn ? Ot.set(e, t) : Ot.has(e) || Ot.set(e, e.v);
		var r = bt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && st(t), pt === null && Je(t);
		}
		e.wv = Rn(), Mt = null, Nt = 0, Lt(e, b, n), Mt = null, Ue() && U !== null && U.f & 1024 && !(U.f & 96) && (Mn === null ? Nn([e]) : Mn.push(e)), !r.is_fork && Dt.size > 0 && !kt && Ft();
	}
	return t;
}
function Ft() {
	kt = !1;
	for (let e of Dt) {
		e.f & 1024 && j(e, x);
		let t;
		try {
			t = zn(e);
		} catch {
			t = !0;
		}
		t && Gn(e);
	}
	Dt.clear();
}
function It(e) {
	F(e, e.v + 1);
}
function Lt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ue(), a = r.length;
		if (Nt += a, Nt > 1e5 && Mt === null && (Mt = /* @__PURE__ */ new Set()), Mt !== null) {
			if (Mt.has(e)) return;
			Mt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== U) {
				var l = (c & b) === 0;
				if (l && j(s, t), c & 131072) Dt.add(s);
				else if (c & 2) {
					var u = s;
					pt?.delete(u), Lt(u, x, n);
				} else if (l) {
					var d = s;
					c & 16 && St !== null && St.add(d), n === null ? wt(d) : n.push(d);
				}
			}
		}
	}
}
function Rt(e) {
	if (typeof e != "object" || !e || ce in e || le in e) return e;
	let n = m(e);
	if (n !== f && n !== p) return e;
	var r = /* @__PURE__ */ new Map(), i = a(e), o = /* @__PURE__ */ P(0), s = null, c = In, l = (e) => {
		if (In === c) return e();
		var t = H, n = In;
		En(null), Ln(c);
		var r = e();
		return En(t), Ln(n), r;
	};
	return i && r.set("length", /* @__PURE__ */ P(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Pe();
			var i = r.get(t);
			return i === void 0 ? l(() => {
				var e = /* @__PURE__ */ P(n.value, s);
				return r.set(t, e), e;
			}) : F(i, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var i = r.get(n);
			if (i === void 0) {
				if (n in e) {
					let e = l(() => /* @__PURE__ */ P(t, s));
					r.set(n, e), It(o);
				}
			} else F(i, t), It(o);
			return !0;
		},
		get(n, i, a) {
			if (i === ce) return e;
			var o = r.get(i), c = i in n;
			if (o === void 0 && (!c || u(n, i)?.writable) && (o = l(() => /* @__PURE__ */ P(Rt(c ? n[i] : t), s)), r.set(i, o)), o !== void 0) {
				var d = W(o);
				return d === t ? void 0 : d;
			}
			return Reflect.get(n, i, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var i = Reflect.getOwnPropertyDescriptor(e, n), a = r.get(n);
			if (a !== void 0) {
				var o = W(a);
				if (o === t) return;
				if (i && "value" in i) i.value = o;
				else return {
					enumerable: !0,
					configurable: !0,
					value: o,
					writable: !0
				};
			}
			return i;
		},
		has(e, n) {
			if (n === ce) return !0;
			var i = r.get(n), a = i !== void 0 && i.v !== t || Reflect.has(e, n);
			return (i !== void 0 || (U !== null || H !== null) && (!a || u(e, n)?.writable)) && (i === void 0 && (i = l(() => /* @__PURE__ */ P(a ? Rt(e[n]) : t, s)), r.set(n, i)), W(i) === t) ? !1 : a;
		},
		set(e, n, a, c) {
			var d = r.get(n), f = n in e;
			if (i && n === "length") for (var p = a; p < d.v; p += 1) {
				var m = r.get(p + "");
				m === void 0 ? p in e && (m = l(() => /* @__PURE__ */ P(t, s)), r.set(p + "", m)) : F(m, t);
			}
			if (d === void 0) (!f || u(e, n)?.writable) && (d = l(() => /* @__PURE__ */ P(void 0, s)), F(d, Rt(a)), r.set(n, d));
			else {
				f = d.v !== t;
				var h = l(() => Rt(a));
				F(d, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, a), !f) {
				if (i && typeof n == "string") {
					var _ = r.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && F(_, v + 1);
				}
				It(o);
			}
			return !0;
		},
		ownKeys(e) {
			W(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = r.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [i, a] of r) a.v !== t && !(i in e) && n.push(i);
			return n;
		},
		setPrototypeOf() {
			Fe();
		}
	});
}
var zt, Bt, Vt, Ht;
function Ut() {
	if (zt === void 0) {
		zt = window, Bt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Vt = u(t, "firstChild").get, Ht = u(t, "nextSibling").get, h(e) && (e[pe] = void 0, e[fe] = null, e[me] = void 0, e.__e = void 0), h(n) && (n[he] = void 0);
	}
}
function I(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function L(e) {
	return Vt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Wt(e) {
	return Ht.call(e);
}
function R(e, t) {
	if (!E) return /* @__PURE__ */ L(e);
	var n = /* @__PURE__ */ L(D);
	if (n === null) n = D.appendChild(I());
	else if (t && n.nodeType !== 3) {
		var r = I();
		return n?.before(r), xe(r), r;
	}
	return t && Yt(n), xe(n), n;
}
function Gt(e, t = !1) {
	if (!E) {
		var n = /* @__PURE__ */ L(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Wt(n) : n;
	}
	if (t) {
		if (D?.nodeType !== 3) {
			var r = I();
			return D?.before(r), xe(r), r;
		}
		Yt(D);
	}
	return D;
}
function z(e, t = !1) {
	if (!E) return /* @__PURE__ */ L(e);
	var n = R(e, t);
	return O(e), n;
}
function B(e, t = 1, n = !1) {
	let r = E ? D : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Wt(r);
	if (!E) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = I();
			return r === null ? i?.after(a) : r.before(a), xe(a), a;
		}
		Yt(r);
	}
	return xe(r), r;
}
function Kt(e) {
	e.textContent = "";
}
function qt() {
	return !1;
}
function Jt(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function Yt(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function Xt(e) {
	var t = U;
	if (t === null) return H.f |= se, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	Zt(e, t);
}
function Zt(t, n) {
	if (t === e) throw t;
	if (!(n !== null && n.f & 16384)) {
		for (; n !== null;) {
			if (n.f & 128 && !(n.f & 33570816)) {
				if (!(n.f & 32768)) throw t;
				try {
					n.b.error(t);
					return;
				} catch (e) {
					t = e;
				}
			}
			n = n.parent;
		}
		throw t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/effects.js
function Qt(e) {
	U === null && (H === null && je(e), Ae()), Cn && ke(e);
}
function $t(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function en(e, t) {
	var n = U;
	n !== null && n.f & 8192 && (e |= S);
	var r = {
		ctx: Re,
		deps: null,
		nodes: null,
		f: e | b | 512,
		first: null,
		fn: t,
		last: null,
		next: null,
		parent: n,
		b: n && n.b,
		prev: null,
		teardown: null,
		wv: 0,
		ac: null
	};
	M?.register_created_effect(r);
	var i = r;
	if (e & 4) gt === null ? bt.ensure().schedule(r) : gt.push(r);
	else if (t !== null) {
		try {
			Gn(r);
		} catch (e) {
			throw mn(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= ne));
	}
	if (i !== null && (i.parent = n, n !== null && $t(i, n), H !== null && H.f & 2 && !(e & 64))) {
		var a = H;
		(a.effects ??= []).push(i);
	}
	return r;
}
function tn() {
	return H !== null && !Tn;
}
function nn(e) {
	let t = en(8, null);
	return j(t, y), t.teardown = e, t;
}
function rn(e) {
	Qt("$effect");
	var t = U.f;
	if (!H && t & 32 && Re !== null && !Re.i) {
		var n = Re;
		(n.e ??= []).push(e);
	} else return an(e);
}
function an(e) {
	return en(4 | re, e);
}
function on(e) {
	bt.ensure();
	let t = en(64 | w, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? _n(t, () => {
			mn(t), n(void 0);
		}) : (mn(t), n(void 0));
	});
}
function sn(e) {
	return en(oe | w, e);
}
function cn(e, t = 0) {
	return en(8 | t, e);
}
function V(e, t = [], n = [], r = []) {
	Ze(r, t, n, (t) => {
		en(8, () => {
			e(...t.map(W));
		});
	});
}
function ln(e, t = 0) {
	return en(16 | t, e);
}
function un(e) {
	return en(32 | w, e);
}
function dn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = Cn, r = H;
		wn(!0), En(null);
		try {
			t.call(null);
		} catch (t) {
			Zt(t, e.parent);
		} finally {
			wn(n), En(r);
		}
	}
}
function fn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Xe(() => {
			e.abort(ge);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : mn(n, t), n = r;
	}
}
function pn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || mn(t), t = n;
	}
}
function mn(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (hn(e.nodes.start, e.nodes.end), n = !0), e.f |= te, fn(e, t && !n), Wn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	dn(e), e.f ^= te, e.f |= C;
	var i = e.parent;
	i !== null && i.first !== null && gn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function hn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Wt(e);
		e.remove(), e = n;
	}
}
function gn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function _n(e, t, n = !0) {
	var r = [];
	e.f |= 256, vn(e, r, !0);
	var i = () => {
		n && mn(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function vn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= S;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				vn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function yn(e) {
	e.f &= -257, bn(e, !0);
}
function bn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= S, e.f & 1024 || (j(e, b), bt.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			bn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function xn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ Wt(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var Sn = null, Cn = !1;
function wn(e) {
	Cn = e;
}
var H = null, Tn = !1;
function En(e) {
	H = e;
}
var U = null;
function Dn(e) {
	U = e;
}
var On = null;
function kn(e) {
	H !== null && (H.f & 2097152 || H.f & 2) && (On ??= /* @__PURE__ */ new Set()).add(e);
}
var An = null, jn = 0, Mn = null;
function Nn(e) {
	Mn = e;
}
var Pn = 1, Fn = 0, In = Fn;
function Ln(e) {
	In = e;
}
function Rn() {
	return ++Pn;
}
function zn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (zn(a) && ct(a), a.wv > e.wv) return !0;
		}
		t & 512 && pt === null && j(e, y);
	}
	return !1;
}
function Bn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(On !== null && On.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Bn(a, t, !1) : t === a && (n ? j(a, b) : a.f & 1024 && j(a, x), wt(a));
	}
}
function Vn(e) {
	var t = An, n = jn, r = Mn, i = H, a = On, o = Re, s = Tn, c = In, l = e.f;
	An = null, jn = 0, Mn = null, H = l & 96 ? null : e, On = null, ze(e.ctx), Tn = !1, In = ++Fn, e.ac !== null && (Xe(() => {
		e.ac.abort(ge);
	}), e.ac = null);
	try {
		e.f |= ae;
		var u = e.fn, d = u();
		e.f |= ee;
		var f = Hn(e);
		if (Ue() && Mn !== null && !Tn && f !== null && !(e.f & 6146)) for (var p = 0; p < Mn.length; p++) Bn(Mn[p], e);
		if (i !== null && i !== e) {
			if (Fn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Fn;
			if (t !== null) for (let e of t) e.rv = Fn;
			Mn !== null && (r === null ? r = Mn : r.push(...Mn));
		}
		return e.f & 8388608 && (e.f ^= se), d;
	} catch (t) {
		return Hn(e), Xt(t);
	} finally {
		e.f ^= ae, An = t, jn = n, Mn = r, H = i, On = a, ze(o), Tn = s, In = c;
	}
}
function Hn(e) {
	var t = e.deps, n = M?.is_fork;
	if (An !== null) {
		var r;
		if (n || Wn(e, jn), t !== null && jn > 0) for (t.length = jn + An.length, r = 0; r < An.length; r++) t[jn + r] = An[r];
		else e.deps = t = An;
		if (tn() && e.f & 512) for (r = jn; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && jn < t.length && (Wn(e, jn), t.length = jn);
	return t;
}
function Un(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var i = o.call(r, e);
		if (i !== -1) {
			var a = r.length - 1;
			a === 0 ? r = n.reactions = null : (r[i] = r[a], r.pop());
		}
	}
	if (r === null && n.f & 2 && (An === null || !s.call(An, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Je(c), c.ac !== null && Xe(() => {
			c.ac.abort(ge), c.ac = null, j(c, b);
		}), lt(c), Wn(c, 0);
	}
}
function Wn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Un(e, n[r]);
}
function Gn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		j(e, y);
		var n = U;
		U = e;
		try {
			t & 16777232 ? pn(e) : fn(e), dn(e);
			var r = Vn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Pn;
		} finally {
			U = n;
		}
	}
}
function W(e) {
	var t = !!(e.f & 2);
	if (Sn?.add(e), H !== null && !Tn && !(U !== null && U.f & 16384) && (On === null || !On.has(e))) {
		var n = H.deps;
		if (H.f & 2097152) e.rv < Fn && (e.rv = Fn, An === null && n !== null && n[jn] === e ? jn++ : An === null ? An = [e] : An.push(e));
		else {
			H.deps ??= [], s.call(H.deps, e) || H.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [H] : s.call(r, H) || r.push(H);
		}
	}
	if (Cn && Ot.has(e)) return Ot.get(e);
	if (t) {
		var i = e;
		if (Cn) {
			var a = i.v;
			return (!(i.f & 1024) && i.reactions !== null || qn(i)) && (a = st(i)), Ot.set(i, a), a;
		}
		var o = !(i.f & 512) && !Tn && H !== null && !!(H.f & 512), c = (i.f & ee) === 0;
		zn(i) && (o && (i.f |= 512), ct(i)), o && !c && (ut(i), Kn(i));
	}
	if (pt?.has(e)) return pt.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function Kn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : s.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (ut(n), Kn(n));
	}
}
function qn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Ot.has(t) || t.f & 2 && qn(t)) return !0;
	return !1;
}
function Jn(e) {
	var t = Tn;
	try {
		return Tn = !0, e();
	} finally {
		Tn = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var Yn = Symbol("events"), Xn = /* @__PURE__ */ new Set(), Zn = /* @__PURE__ */ new Set();
function Qn(e, t, n) {
	(t[Yn] ??= {})[e] = n;
}
function $n(e) {
	for (var t = 0; t < e.length; t++) Xn.add(e[t]);
	for (var n of Zn) n(e);
}
var er = null, tr = !1;
function nr(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	er = e, tr || (tr = !0, setTimeout(() => {
		tr = !1, er = null;
	}));
	var o = 0, s = er === e && e[Yn];
	if (s) {
		var c = i.indexOf(s);
		if (c !== -1 && (t === document || t === window)) {
			e[Yn] = t;
			return;
		}
		var u = i.indexOf(t);
		if (u === -1) return;
		c <= u && (o = c);
	}
	if (a = i[o] || e.target, a !== t) {
		l(e, "currentTarget", {
			configurable: !0,
			get() {
				return a || n;
			}
		});
		var d = H, f = U;
		En(null), Dn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Yn]?.[r];
					h == null || a.disabled && e.target !== a || h.call(a, e);
				} catch (e) {
					p ? m.push(e) : p = e;
				}
				if (e.cancelBubble) break;
				o++, a = o < i.length ? i[o] : null;
			}
			if (p) {
				for (let e of m) queueMicrotask(() => {
					throw e;
				});
				throw p;
			}
		} finally {
			e[Yn] = t, delete e.currentTarget, En(d), Dn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var rr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function ir(e) {
	return rr?.createHTML(e) ?? e;
}
function ar(e) {
	var t = Jt("template");
	return t.innerHTML = ir(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function or(e, t) {
	var n = U;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function G(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (E) return or(D, null), D;
		i === void 0 && (i = ar(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ L(i)));
		var t = r || Bt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ L(t), s = t.lastChild;
			or(o, s);
		} else or(t, t);
		return t;
	};
}
function sr(e = "") {
	if (!E) {
		var t = I(e + "");
		return or(t, t), t;
	}
	var n = D;
	return n.nodeType === 3 ? Yt(n) : (n.before(n = I()), xe(n)), or(n, n), n;
}
function cr() {
	if (E) return or(D, null), D;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = I();
	return e.append(t, n), or(t, n), e;
}
function K(e, t) {
	if (E) {
		var n = U;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = D), Se();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var lr = ["touchstart", "touchmove"];
function ur(e) {
	return lr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function dr(e) {
	let t = 0, n = At(0), r;
	return () => {
		tn() && (W(n), cn(() => (t === 0 && (r = Jn(() => e(() => It(n)))), t += 1, () => {
			Ke(() => {
				--t, t === 0 && (r?.(), r = void 0, It(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var fr = ne | w;
function pr(e, t, n, r) {
	new mr(e, t, n, r);
}
var mr = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = E ? D : null;
	#n;
	#r;
	#i;
	#a = null;
	#o = null;
	#s = null;
	#c = null;
	#l = 0;
	#u = 0;
	#d = !1;
	#f = /* @__PURE__ */ new Map();
	#p = null;
	#m = dr(() => (this.#p = At(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = U;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = U.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = ln(() => {
			if (E) {
				let e = A(this.#t);
				Se();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, fr), E && (this.#e = D);
	}
	#h() {
		try {
			this.#a = un(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		Ke(r), t && (this.#s = un(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? ye() : (t = !0, n && Le(), this.#s !== null && _n(this.#s, () => {
				this.#s = null;
			}), this.#S(() => {
				this.#b();
			})));
		};
		return {
			reset: r,
			invoke_onerror: () => {
				if (!this.#v()) try {
					n = !0, this.#n.onerror?.(e, r), n = !1;
				} catch (e) {
					Zt(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (C | te)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = un(() => e(this.#e)), Ke(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = I(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return un(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							Zt(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(M)) : this.#u === 0 && (this.#e.before(e), this.#c = null, _n(this.#o, () => {
					this.#o = null;
				}), this.#x(M));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = un(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				xn(this.#a, e);
				let t = this.#n.pending;
				this.#o = un(() => t(this.#e));
			} else this.#x(M);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		Ye(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = U, n = H, r = Re;
		Dn(this.#i), En(this.#i), ze(this.#i.ctx);
		try {
			return bt.ensure(), e();
		} finally {
			Dn(t), En(n), ze(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && _n(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Ke(() => {
			this.#d = !1, this.#p && Pt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), W(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		M?.is_fork ? (this.#a && M.skip_effect(this.#a), this.#o && M.skip_effect(this.#o), this.#s && M.skip_effect(this.#s), M.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (mn(this.#a), null), this.#o &&= (mn(this.#o), null), this.#s &&= (mn(this.#s), null), E && (xe(this.#t), k(), xe(Ce()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return un(() => {
						var r = U;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return Zt(e, this.#i.parent), null;
				}
			}));
		};
		Ke(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					Zt(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => Zt(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function q(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[he] ??= e.nodeValue) && (e[he] = n, e.nodeValue = `${n}`);
}
function hr(e, t) {
	return _r(e, t);
}
var gr = /* @__PURE__ */ new Map();
function _r(t, { target: n, anchor: r, props: i = {}, events: a, context: o, intro: s = !0, transformError: l }) {
	Ut();
	var u = void 0, d = on(() => {
		var s = r ?? n.appendChild(I());
		pr(s, { pending: () => {} }, (n) => {
			Be({});
			var r = Re;
			if (o && (r.c = o), a && (i.$$events = a), E && or(n, null), u = t(n, i) || He(), E && (U.nodes.end = D, D === null || D.nodeType !== 8 || D.data !== "]")) throw ve(), e;
			Ve();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = ur(r);
					for (let e of [n, document]) {
						var a = gr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), gr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, nr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(c(Xn)), Zn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = gr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, nr), t.delete(e), t.size === 0 && gr.delete(r)) : t.set(e, i);
			}
			Zn.delete(f), s !== r && s.parentNode?.removeChild(s);
		};
	});
	return vr.set(u, d), u;
}
var vr = /* @__PURE__ */ new WeakMap();
function yr(e, t) {
	let n = vr.get(e);
	return n ? (vr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var br = class {
	anchor;
	#e = /* @__PURE__ */ new Map();
	#t = /* @__PURE__ */ new Map();
	#n = /* @__PURE__ */ new Map();
	#r = /* @__PURE__ */ new Set();
	#i = !0;
	constructor(e, t = !0) {
		this.anchor = e, this.#i = t;
	}
	#a = (e) => {
		if (this.#e.has(e)) {
			var t = this.#e.get(e), n = this.#t.get(t);
			if (n) yn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (yn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (mn(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						xn(r, t), t.append(I()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else mn(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), _n(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (mn(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = M, r = qt();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = I();
				i.append(a), this.#n.set(e, {
					effect: un(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, un(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else E && (this.anchor = D), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function J(e, t, n = !1) {
	var r;
	E && (r = D, Se());
	var i = new br(e), a = n ? ne : 0;
	function o(e, t) {
		if (E) {
			var n = A(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Ce();
				xe(a), i.anchor = a, be(!1), i.ensure(e, t), be(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	ln(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function xr(e, t, n) {
	for (var r = [], i = t.length, a, o = t.length, s = 0; s < i; s++) {
		let n = t[s];
		_n(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					Sr(e, c(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --o;
		}, !1);
	}
	if (o === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			Kt(d), d.append(u), e.items.clear();
		}
		Sr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function Sr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= ie, xn(a, document.createDocumentFragment())) : mn(t[i], n);
	}
}
var Cr;
function wr(e, t, n, r, i, o = null) {
	var s = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		s = E ? xe(/* @__PURE__ */ L(u)) : u.appendChild(I());
	}
	E && Se();
	var d = null, f = /* @__PURE__ */ at(() => {
		var e = n();
		return a(e) ? e : e == null ? [] : c(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = W(f);
			_.fallback = d, Er(_, n, s, t, r), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= ie, Or(d, null, s)) : yn(d) : _n(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: ln(() => {
			var e = W(f), a = e.length;
			let c = !1;
			E && A(s) === "[!" != (a === 0) && (s = Ce(), xe(s), be(!1), c = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = M, v = qt(), y = 0; y < a; y += 1) {
				E && D.nodeType === 8 && D.data === "]" && (s = D, c = !0, be(!1));
				var b = e[y], x = r(b, y), S = m ? null : l.get(x);
				S ? (S.v && Pt(S.v, b), S.i && Pt(S.i, y), v && _.unskip_effect(S.e)) : (S = Dr(l, m ? s : Cr ??= I(), b, x, y, i, t, n), m || (S.e.f |= ie), l.set(x, S)), u.add(x);
			}
			if (a === 0 && o && !d && (m ? d = un(() => o(s)) : (d = un(() => o(Cr ??= I())), d.f |= ie)), a > u.size && Oe("", "", ""), E && a > 0 && xe(Ce()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			c && be(!0), W(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, E && (s = D);
}
function Tr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function Er(e, t, n, r, i) {
	var a = !!(r & 8), o = t.length, s = e.items, l = Tr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < o; v += 1) h = t[v], g = i(h, v), _ = s.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < o; v += 1) {
		if (h = t[v], g = i(h, v), _ = s.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (yn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= ie, _ === l) Or(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), kr(e, d, _), kr(e, _, y), Or(_, y, n), d = _, p = [], m = [], l = Tr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) Or(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					kr(e, S.prev, C.next), kr(e, d, S), kr(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), Or(_, l, n), kr(e, _.prev, _.next), kr(e, _, d === null ? e.effect.first : d.next), kr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = Tr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = Tr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (Sr(e, c(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var ee = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || ee.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && ee.push(l), l = Tr(l.next);
		var te = ee.length;
		if (te > 0) {
			var ne = r & 4 && o === 0 ? n : null;
			if (a) {
				for (v = 0; v < te; v += 1) ee[v].nodes?.a?.measure();
				for (v = 0; v < te; v += 1) ee[v].nodes?.a?.fix();
			}
			xr(e, ee, ne);
		}
	}
	a && Ke(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Dr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? At(n) : /* @__PURE__ */ jt(n, !1, !1) : null, l = o & 2 ? At(i) : null;
	return {
		v: c,
		i: l,
		e: un(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Or(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Wt(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function kr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
function Ar(t, n, a = !1, o = !1, s = !1, c = !1) {
	var l = t, u = "";
	if (a) {
		var d = t;
		E && (l = xe(/* @__PURE__ */ L(d)));
	}
	V(() => {
		var t = U;
		if (u === (u = n() ?? "")) E && Se();
		else if (a && !E) t.nodes = null, d.innerHTML = u, u !== "" && or(/* @__PURE__ */ L(d), d.lastChild);
		else if (t.nodes !== null && (hn(t.nodes.start, t.nodes.end), t.nodes = null), u !== "") {
			if (E) {
				for (var c = D.data, f = Se(), p = f; f !== null && (f.nodeType !== 8 || f.data !== "");) p = f, f = /* @__PURE__ */ Wt(f);
				if (f === null) throw ve(), e;
				or(D, p), l = xe(f);
			} else {
				var m = Jt(o ? "svg" : s ? "math" : "template", o ? r : s ? i : void 0);
				m.innerHTML = u;
				var h = o || s ? m : m.content;
				if (or(/* @__PURE__ */ L(h), h.lastChild), o || s) for (; /* @__PURE__ */ L(h);) l.before(/* @__PURE__ */ L(h));
				else l.before(h);
			}
		}
	});
}
// node_modules/svelte/src/internal/shared/attributes.js
var jr = [..." 	\n\r\f\xA0\v﻿"];
function Mr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || jr.includes(r[o - 1])) && (s === r.length || jr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Nr(e, t, n, r, i, a) {
	var o = e[pe];
	if (E || o !== n || o === void 0) {
		var s = Mr(n, r, a);
		E && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[pe] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Pr = Symbol("is custom element"), Fr = Symbol("is html"), Ir = T ? "link" : "LINK";
function Lr(e, t, n, r) {
	var i = Rr(e);
	if (!(E && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Ir))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[de] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Br(e).has(t) ? e[t] = n : (!E || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Rr(e) {
	return e[fe] ??= {
		[Pr]: e.nodeName.includes("-"),
		[Fr]: e.namespaceURI === n
	};
}
var zr = /* @__PURE__ */ new Map();
function Br(e) {
	var t = e.getAttribute("is") || e.nodeName, n = zr.get(t);
	if (n) return n;
	zr.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = d(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = m(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var Vr = !1;
function Hr(e) {
	var t = Vr;
	try {
		return Vr = !1, [e(), Vr];
	} finally {
		Vr = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
function Ur(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, c = !0, l = void 0, d = () => o && i ? (l ??= /* @__PURE__ */ tt(r), W(l)) : (c && (c = !1, s = o ? Jn(r) : r), s);
	let f;
	if (a) {
		var p = ce in e || ue in e;
		f = u(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = Hr(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Ne(t), f(m)));
	var g = i ? () => {
		var n = e[t];
		return n === void 0 ? d() : (c = !0, n);
	} : () => {
		var n = e[t];
		return n !== void 0 && (s = void 0), n === void 0 ? s : n;
	};
	if (i && !(n & 4)) return g;
	if (f) {
		var _ = e.$$legacy;
		return (function(e, t) {
			return arguments.length > 0 ? ((!i || !t || _ || h) && f(t ? g() : e), e) : g();
		});
	}
	var v = !1, y = (n & 1 ? tt : at)(() => (v = !1, g()));
	a && W(y);
	var b = U;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? W(y) : i && a ? Rt(e) : e;
			return F(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return Cn && v || b.f & 16384 ? y.v : W(y);
	});
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// kit/attention.js
function Wr(e, t, n) {
	if (!t) return [];
	let r = (t, n) => () => e.call(t, n), i = () => e.post("api/attention/ack", {
		kind: t,
		id: n
	});
	switch (t) {
		case "run": return [{
			label: "Resume run",
			run: r("run.resume", { id: n }),
			ask: "Resume run " + n + "?"
		}];
		case "pool": return [{
			label: "Unquarantine pool",
			run: r("pool.unquarantine", { pool: n }),
			ask: "Unquarantine pool " + n + "?"
		}];
		case "wait": return [{
			label: "Clear wait",
			run: r("wait.clear", {
				id: n,
				reason: "cleared in Desk"
			}),
			ask: "Clear the wait on " + n + "?"
		}];
		case "decision": return [
			{
				label: "Approve as target",
				run: r("decision.resolve", {
					id: n,
					action: "approve"
				}),
				ask: "Approve " + n + ": relabel the issue proposed to target?"
			},
			{
				label: "Reject",
				run: r("decision.resolve", {
					id: n,
					action: "reject"
				}),
				ask: "Reject " + n + ": close the issue as not planned?"
			},
			{
				label: "Dismiss",
				run: r("decision.resolve", {
					id: n,
					action: "dismiss"
				}),
				ask: "Dismiss " + n + " for now? The issue is left as is."
			}
		];
		case "ticket":
		case "question":
		case "stale": return [{
			label: "Close ticket",
			run: r("ticket.close", { id: n }),
			ask: "Close " + n + " as done?"
		}];
		default: return [{
			label: "Acknowledge",
			run: i,
			ask: "Acknowledge " + t + ":" + n + "? It is hidden until it recurs."
		}];
	}
}
async function Gr(e, t) {
	if (!await e.confirm(t.ask)) return !1;
	try {
		return await t.run(), e.toast(t.label + ": done"), !0;
	} catch (n) {
		return e.toast(t.label + " failed: " + String(n.message || n)), !1;
	}
}
var Kr = Symbol.for("jobshop.desk.attention.pending"), qr = "jobshop-attention-pending";
function Jr() {
	let e = globalThis;
	return e[Kr] ??= /* @__PURE__ */ new Map();
}
function Yr(e, t) {
	Jr().set(e, t), globalThis.dispatchEvent?.(new CustomEvent(qr, { detail: {
		key: e,
		label: t
	} }));
}
function Xr(e) {
	Jr().delete(e);
}
function Zr(e) {
	let t = (t) => e(t.detail.key, t.detail.label);
	return globalThis.addEventListener?.(qr, t), () => globalThis.removeEventListener?.(qr, t);
}
// kit/format.js
function Qr(e) {
	try {
		let t = new URL(e);
		return t.protocol === "https:" || t.protocol === "http:";
	} catch {
		return !1;
	}
}
function $r(e, t) {
	let n = t.title || "", r = t.target || "", i = t.intent || "";
	return e?.name === "jcode" && i ? {
		label: i,
		detail: r || n
	} : n ? {
		label: n,
		detail: r || i
	} : {
		label: r || i || "tool",
		detail: r && i ? i : ""
	};
}
// node_modules/marked/lib/marked.esm.js
function ei() {
	return {
		async: !1,
		breaks: !1,
		extensions: null,
		gfm: !0,
		hooks: null,
		pedantic: !1,
		renderer: null,
		silent: !1,
		tokenizer: null,
		walkTokens: null
	};
}
var ti = ei();
function ni(e) {
	ti = e;
}
var ri = { exec: () => null };
function ii(e) {
	let t = [];
	return (n) => {
		let r = Math.max(0, Math.min(3, n - 1)), i = t[r];
		return i || (i = e(r), t[r] = i), i;
	};
}
function Y(e, t = "") {
	let n = typeof e == "string" ? e : e.source, r = {
		replace: (e, t) => {
			let i = typeof t == "string" ? t : t.source;
			return i = i.replace(oi.caret, "$1"), n = n.replace(e, i), r;
		},
		getRegex: () => new RegExp(n, t)
	};
	return r;
}
var ai = ((e = "") => {
	try {
		return !!RegExp("(?<=1)(?<!1)" + e);
	} catch {
		return !1;
	}
})(), oi = {
	codeRemoveIndent: /^(?: {0,3}\t| {1,4})/gm,
	outputLinkReplace: /\\([\[\]])/g,
	indentCodeCompensation: /^(\s+)(?:```)/,
	beginningSpace: /^\s+/,
	endingHash: /#$/,
	startingSpaceChar: /^ /,
	endingSpaceChar: / $/,
	endingSpaceTabChar: /[ \t]$/,
	nonSpaceChar: /[^ ]/,
	newLineCharGlobal: /\n/g,
	tabCharGlobal: /\t/g,
	leadingSpaceTab: /^[ \t]+/,
	multipleSpaceGlobal: /\s+/g,
	blankLine: /^[ \t]*$/,
	doubleBlankLine: /\n[ \t]*\n[ \t]*$/,
	blockquoteStart: /^ {0,3}>/,
	blockquoteSetextReplace: /\n {0,3}((?:=+|-+)[ \t]*)(?=\n|$)/g,
	blockquoteSetextReplace2: /^ {0,3}>[ \t]?/gm,
	blockquoteMarkerWhitespace: /^( {0,3}>)([ \t]+)/gm,
	listReplaceNesting: /^ {1,4}(?=( {4})*[^ ])/g,
	listIsTask: /^\[[ xX]\][ \t]+\S/,
	listReplaceTask: /^\[[ xX]\][ \t]+/,
	listTaskCheckbox: /\[[ xX]\]/,
	anyLine: /\n.*\n/,
	hrefBrackets: /^<(.*)>$/,
	tableDelimiter: /[:|]/,
	tableAlignChars: /^\||\| *$/g,
	tableRowBlankLine: /\n[ \t]*$/,
	tableAlignRight: /^ *-+: *$/,
	tableAlignCenter: /^ *:-+: *$/,
	tableAlignLeft: /^ *:-+ *$/,
	startATag: /^<a /i,
	endATag: /^<\/a>/i,
	startPreScriptTag: /^<(pre|code|kbd|script)(\s|>)/i,
	endPreScriptTag: /^<\/(pre|code|kbd|script)(\s|>)/i,
	startAngleBracket: /^</,
	endAngleBracket: />$/,
	pedanticHrefTitle: /^([^'"]*[^\s])\s+(['"])(.*)\2/,
	unicodeAlphaNumeric: /[\p{L}\p{N}]/u,
	numericCharacterReference: /&#(?:(\d{1,7})|[Xx]([A-Fa-f0-9]{1,6}));/,
	numericCharacterReferenceOrAmpersand: /&#(?:(\d{1,7})|[Xx]([A-Fa-f0-9]{1,6}));|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/g,
	escapeTest: /[&<>"']/,
	escapeReplace: /[&<>"']/g,
	escapeTestNoEncode: /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/,
	escapeReplaceNoEncode: /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/g,
	caret: /(^|[^\[])\^/g,
	percentDecode: /%25/g,
	findPipe: /\|/g,
	splitPipe: / \|/,
	slashPipe: /\\\|/g,
	carriageReturn: /\r\n|\r/g,
	spaceLine: /^ +$/gm,
	notSpaceStart: /^\S*/,
	endingNewline: /\n$/,
	listItemRegex: (e) => RegExp(`^( {0,3}${e})((?:[	 ][^\\n]*)?(?:\\n|$))`),
	nextBulletRegex: ii((e) => RegExp(`^ {0,${e}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`)),
	hrRegex: ii((e) => RegExp(`^ {0,${e}}((?:-[ 	]*){3,}|(?:_[ 	]*){3,}|(?:\\*[ 	]*){3,})(?:\\n+|$)`)),
	fencesBeginRegex: ii((e) => RegExp(`^ {0,${e}}(?:\`\`\`|~~~)`)),
	headingBeginRegex: ii((e) => RegExp(`^ {0,${e}}#{1,6}(?:\\s|$)`)),
	pedanticHeadingBeginRegex: ii((e) => RegExp(`^ {0,${e}}#`)),
	htmlBeginRegex: ii((e) => RegExp(`^ {0,${e}}(?:</?(?:${xi})(?: +|$|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)`, "i")),
	blockquoteBeginRegex: ii((e) => RegExp(`^ {0,${e}}>`))
}, si = /^(?:[ \t]*(?:\n|$))+/, ci = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, li = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}(?:(?=`)\1`*|(?=~)\1~*) *(?=\n|$)|$)/, ui = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, di = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, fi = / {0,3}(?:[*+-]|\d{1,9}[.)])/, pi = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |fences|blockquote|heading|hr|html|table))+?)\n {0,3}(=+|-+)[ \t]*(?:\n+|$)/, mi = Y(pi).replace(/bull/g, fi).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), hi = Y(pi).replace(/bull/g, fi).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), gi = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table|[ \t]+\n)[^\n]+)*)/, _i = /^[^\n]+/, vi = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\\n]|\n(?![ \t]*\n))+/, yi = Y(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", vi).replace("title", /(?:"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\\n]|\n(?!\n))*'|\((?:\\[\s\S]|[^()\\])*\))/).getRegex(), bi = Y(/^(bull)([ \t][^\n]*?)?(?:\n|$)/).replace(/bull/g, fi).getRegex(), xi = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", Si = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, Ci = Y("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n*|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>[^\\n]*\\n*|$)|<![A-Z][\\s\\S]*?(?:>[^\\n]*\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>[^\\n]*\\n*|$)|</?(tag)(?:[ \\t]+|\\n|/?>|$)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!(?:script|pre|style|textarea)(?![a-z0-9-]))([a-z][a-z0-9-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</[a-z][a-z0-9-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", Si).replace("tag", xi).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), wi = (e) => Y(gi).replace("hr", ui).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", e).replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", xi).getRegex(), Ti = wi(/ {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]/), Ei = wi(/ {0,3}(?:[*+-]|\d{1,9}[.)])(?:[ \t]|\n|$)/), Di = {
	blockquote: Y(/^( {0,3}> ?(?:(?![ \t]*(?:\n|$))(paragraph)|[^\n]*)(?:\n|$))+/).replace("paragraph", Ei).getRegex(),
	code: ci,
	def: yi,
	fences: li,
	heading: di,
	hr: ui,
	html: Ci,
	lheading: mi,
	list: bi,
	newline: si,
	paragraph: Ti,
	table: ri,
	text: _i
}, Oi = Y("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", ui).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", xi).getRegex(), ki = {
	...Di,
	lheading: hi,
	table: Oi,
	paragraph: Y(gi).replace("hr", ui).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", Oi).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]+[^ \\t\\n]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", xi).getRegex()
}, Ai = {
	...Di,
	html: Y("^ *(?:comment(?: *\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1>(?: *\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?>(?: *\\n{2,}|\\s*$))").replace("comment", Si).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: ri,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: Y(gi).replace("hr", ui).replace("heading", " *#{1,6} *[^\n]").replace("lheading", mi).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, ji = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, Mi = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, Ni = /^( {2,}|\\)\n(?!\s*$)[ \t]*/, Pi = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, Fi = /[\p{P}\p{S}]/u, Ii = /[\s\p{P}\p{S}]/u, Li = /[^\s\p{P}\p{S}]/u, Ri = Y(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, Ii).getRegex(), zi = /[\p{Pi}\p{Ps}"']/u, Bi = /(?!~)[\p{P}\p{S}]/u, Vi = /(?!~)[\s\p{P}\p{S}]/u, Hi = /(?:[^\s\p{P}\p{S}]|~)/u, Ui = Y(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", ai ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), Wi = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, Gi = Y(Wi, "u").replace(/punct/g, Fi).getRegex(), Ki = Y(Wi, "u").replace(/punct/g, Bi).getRegex(), qi = Y(/^(?:\*+(?:((?!\*)(?!openQuote)punct)|([^\s*]))?)|^_+(?:((?!_)(?!openQuote)punct)|([^\s_]))?/, "u").replace(/openQuote/g, zi).replace(/punct/g, Fi).getRegex(), Ji = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", Yi = Y(Ji, "gu").replace(/notPunctSpace/g, Li).replace(/punctSpace/g, Ii).replace(/punct/g, Fi).getRegex(), Xi = Y(Ji, "gu").replace(/notPunctSpace/g, Hi).replace(/punctSpace/g, Vi).replace(/punct/g, Bi).getRegex(), Zi = Y("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)[\\s](\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|(?:(?!\\*)punct|notPunctSpace)(\\*+)(?!\\*)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Li).replace(/punctSpace/g, Ii).replace(/punct/g, Fi).getRegex(), Qi = Y("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, Li).replace(/punctSpace/g, Ii).replace(/punct/g, Fi).getRegex(), $i = Y("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)[\\s](_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)|(?:(?!_)punct|notPunctSpace)(_+)(?!_)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Li).replace(/punctSpace/g, Ii).replace(/punct/g, Fi).getRegex(), ea = Y(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, Fi).getRegex(), ta = Y("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Li).replace(/punctSpace/g, Ii).replace(/punct/g, Fi).getRegex(), na = Y(/\\(punct)/, "gu").replace(/punct/g, Fi).getRegex(), ra = Y(/(`+)(?:[^`]|[^`][\s\S]*?[^`])\1(?!`)|`+|\\(punct)/, "gu").replace(/punct/g, Fi).getRegex(), ia = Y(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), aa = Y(Si).replace("(?:-->|$)", "-->").getRegex(), oa = Y("^comment|^</[a-zA-Z][a-zA-Z0-9-]*\\s*>|^<[a-zA-Z][a-zA-Z0-9-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", aa).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), sa = /\[(?:\\[\s\S]|[^\[\]\\])*\]/, ca = Y(/(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/).replace("brackets", sa).getRegex(), la = Y(/^!?\[(label)\]\([ \t\n]*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?[ \t\n]*\)/).replace("label", ca).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]+|(?=\))/).replace("title", /"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|\((?:\\[\s\S]|[^)\\])*\)/).getRegex(), ua = Y(/^!?\[(label)\]\[(ref)\]/).replace("label", ca).replace("ref", vi).getRegex(), da = Y(/^!?\[(ref)\](?:\[\])?/).replace("ref", vi).getRegex(), fa = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\]){1,999}/, pa = Y(/(?:[^\[\]\\`]*(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\]))){0,999}?[^\[\]\\`]*?/).replace("brackets", sa).getRegex(), ma = Y("reflink|nolink(?!\\()", "g").replace("reflink", Y(/^!?\[(label)\]\[(ref)\]/).replace("label", pa).replace("ref", fa).getRegex()).replace("nolink", Y(/^!?\[(ref)\](?:\[\])?/).replace("ref", fa).getRegex()).getRegex(), ha = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, ga = Y(/(?:mailto:email|xmpp:email(?:\/[A-Za-z0-9@.]+)?)/).replace(/email/g, /[A-Za-z0-9._+-]+@[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(), _a = {
	_backpedal: ri,
	anyPunctuation: na,
	anyPunctuationOutsideCode: ra,
	autolink: ia,
	blockSkip: Ui,
	br: Ni,
	code: Mi,
	del: ri,
	delLDelim: ri,
	delRDelim: ri,
	emStrongLDelim: Gi,
	emStrongRDelimAst: Yi,
	emStrongRDelimUnd: Qi,
	escape: ji,
	link: la,
	nolink: da,
	punctuation: Ri,
	reflink: ua,
	reflinkSearch: ma,
	tag: oa,
	text: Pi,
	url: ri
}, va = {
	..._a,
	emStrongLDelim: qi,
	emStrongRDelimAst: Zi,
	emStrongRDelimUnd: $i,
	link: Y(/^!?\[(label)\]\((.*?)\)/).replace("label", ca).getRegex(),
	reflink: Y(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", ca).getRegex()
}, ya = {
	..._a,
	emStrongRDelimAst: Xi,
	emStrongLDelim: Ki,
	delLDelim: ea,
	delRDelim: ta,
	url: Y(/^emailProtocol|^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("emailProtocol", ga).replace("protocol", ha).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: Y(/^(?:[^a-zA-Z0-9](?=emailProtocol)|(`+|~+|[^`~])(?:(?=[`~])|(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|wwwStart|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9](?=emailProtocol)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))))/).replace("protocol", ha).replace(/emailProtocol/g, /(?:mailto|xmpp):/).replace("wwwStart", ai ? "(?<![^\\s*_~(])www\\." : "www\\.").getRegex()
}, ba = {
	...ya,
	br: Y(Ni).replace("{2,}", "*").getRegex(),
	text: Y(ya.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, xa = {
	normal: Di,
	gfm: ki,
	pedantic: Ai
}, Sa = {
	normal: _a,
	gfm: ya,
	breaks: ba,
	pedantic: va
}, Ca = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, wa = (e) => Ca[e];
function Ta(e, t) {
	if (t) {
		if (oi.escapeTest.test(e)) return e.replace(oi.escapeReplace, wa);
	} else if (oi.escapeTestNoEncode.test(e)) return e.replace(oi.escapeReplaceNoEncode, wa);
	return e;
}
function Ea(e) {
	return oi.numericCharacterReference.test(e) ? e.replace(oi.numericCharacterReferenceOrAmpersand, (e, t, n) => {
		if (t === void 0 && n === void 0) return "&amp;";
		let r = t === void 0 ? Number.parseInt(n, 16) : Number.parseInt(t, 10);
		return r === 0 || r > 1114111 || r >= 55296 && r <= 57343 ? "�" : r === 38 ? "&amp;" : String.fromCodePoint(r);
	}) : e;
}
function Da(e) {
	try {
		e = encodeURI(e).replace(oi.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function Oa(e, t) {
	let n = e.replace(oi.findPipe, (e, t, n) => {
		let r = !1, i = t;
		for (; --i >= 0 && n[i] === "\\";) r = !r;
		return r ? "|" : " |";
	}).split(oi.splitPipe), r = 0;
	if (n[0].trim() || n.shift(), n.length > 0 && !n.at(-1)?.trim() && n.pop(), t) {
		if (n.length > t) n.splice(t);
		else for (; n.length < t;) n.push("");
	}
	for (; r < n.length; r++) n[r] = n[r].trim().replace(oi.slashPipe, "|");
	return n;
}
function ka(e, t, n) {
	let r = e.length;
	if (r === 0) return "";
	let i = 0;
	for (; i < r;) {
		let a = e.charAt(r - i - 1);
		if (a === t && !n) i++;
		else if (a !== t && n) i++;
		else break;
	}
	return e.slice(0, r - i);
}
function Aa(e) {
	let t = e.split("\n"), n = t.length - 1;
	for (; n >= 0 && oi.blankLine.test(t[n]);) n--;
	return t.length - n <= 2 ? e : t.slice(0, n + 1).join("\n");
}
function ja(e) {
	return e.trim().replace(oi.multipleSpaceGlobal, " ").toLowerCase().toUpperCase().toLowerCase();
}
function Ma(e, t) {
	if (e.indexOf(t[0]) === -1 && e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
var Na = /* @__PURE__ */ new WeakMap();
function Pa(e) {
	let t = {
		"*": -1,
		_: -1
	}, n = -1, r = e.length;
	for (let i = 0; i < r; i++) {
		let a = e[i];
		if (a !== "*" && a !== "_" && a !== "~") continue;
		let o = i === 0 ? "\n" : e[i - 1], s = i + 1;
		for (; s < r && e[s] === a;) s++;
		o !== " " && o !== "	" && o !== "\n" && o !== "\r" && o !== "\f" && (a === "~" ? n = i : t[a] = i), i = s - 1;
	}
	return {
		maskedSrc: e,
		lastPossibleEmStrongCloserIndex: t,
		lastPossibleDelCloserIndex: n
	};
}
function Fa(e, t) {
	let n = Pa(t), r = Na.get(e);
	r ? r.push(n) : Na.set(e, [n]);
}
function Ia(e) {
	let t = Na.get(e);
	t && (t.pop(), t.length === 0 && Na.delete(e));
}
function La(e, t, n) {
	let r = Na.get(e)?.at(-1);
	return r && r.maskedSrc === t ? r.lastPossibleEmStrongCloserIndex[n] : Pa(t).lastPossibleEmStrongCloserIndex[n];
}
function Ra(e, t) {
	let n = Na.get(e)?.at(-1);
	return n && n.maskedSrc === t ? n.lastPossibleDelCloserIndex : Pa(t).lastPossibleDelCloserIndex;
}
function za(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function Ba(e, t, n, r, i) {
	let a = t.href, o = t.title || null, s = e[1].replace(i.other.outputLinkReplace, "$1"), c = e[0].charAt(0) === "!", l = r.state.inLink, u = r.state.linkEmitted, d = r.state.inRawBlock;
	r.state.inLink = !0, r.state.linkEmitted = !1;
	let f = r.inlineTokens(s), p = r.state.linkEmitted;
	if (r.state.linkEmitted = u, r.state.inLink = l, !c) {
		if (p) {
			r.state.inRawBlock = d;
			return;
		}
		r.state.linkEmitted = !0;
	}
	return {
		type: c ? "image" : "link",
		raw: n,
		href: a,
		title: o,
		text: s,
		tokens: f
	};
}
function Va(e, t, n) {
	let r = e.match(n.other.indentCodeCompensation);
	if (r === null) return t;
	let i = r[1];
	return t.split("\n").map((e) => {
		let t = e.match(n.other.beginningSpace);
		if (t === null) return e;
		let [r] = t;
		return e.slice(Math.min(r.length, i.length));
	}).join("\n");
}
function Ha(e, t, n, r) {
	if (!t.includes("<")) return !1;
	for (let i = 0; i < t.length; i++) {
		if (t[i] === "\\") {
			i++;
			continue;
		}
		if (t[i] === "`") {
			let e = r.inline.code.exec(t.slice(i));
			if (e) {
				i += e[0].length - 1;
				continue;
			}
		}
		if (t[i] !== "<") continue;
		let a = e.slice(n + i), o = r.inline.tag.exec(a) || r.inline.autolink.exec(a);
		if (o) {
			if (o[0].length > t.length - i) return !0;
			i += o[0].length - 1;
		}
	}
	return !1;
}
var Ua = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || ti;
	}
	space(e) {
		let t = this.rules.block.newline.exec(e);
		if (t && t[0].length > 0) return {
			type: "space",
			raw: t[0]
		};
	}
	code(e) {
		let t = this.rules.block.code.exec(e);
		if (t) {
			let e = this.options.pedantic ? t[0] : Aa(t[0]);
			return {
				type: "code",
				raw: e,
				codeBlockStyle: "indented",
				text: e.replace(this.rules.other.codeRemoveIndent, "")
			};
		}
	}
	fences(e) {
		let t = this.rules.block.fences.exec(e);
		if (t) {
			let e = t[0], n = Va(e, t[3] || "", this.rules);
			return {
				type: "code",
				raw: e,
				lang: t[2] ? t[2].trim().replace(this.rules.inline.anyPunctuation, "$1") : t[2],
				text: n
			};
		}
	}
	heading(e) {
		let t = this.rules.block.heading.exec(e);
		if (t) {
			let e = t[2].trim();
			if (this.rules.other.endingHash.test(e)) {
				let t = ka(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceTabChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: ka(t[0], "\n"),
				depth: t[1].length,
				text: e,
				tokens: this.lexer.inline(e)
			};
		}
	}
	hr(e) {
		let t = this.rules.block.hr.exec(e);
		if (t) return {
			type: "hr",
			raw: ka(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = ka(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteMarkerWhitespace, (e, t, n) => t + za(n, t.length)).replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
				n = n ? `${n}
${s}` : s, r = r ? `${r}
${c}` : c;
				let l = this.lexer.state.top;
				if (this.lexer.state.top = !0, this.lexer.blockTokens(c, i, !0), this.lexer.state.top = l, e.length === 0) break;
				let u = i.at(-1);
				if (u?.type === "code" || u?.type === "heading" || u?.type === "hr" || u?.type === "html" || u?.type === "table") break;
				if (u?.type === "blockquote") {
					let t = u, a = e.join("\n"), o = t.raw + "\n" + a.replace(this.rules.other.blockquoteSetextReplace2, ""), s = this.blockquote(o);
					i[i.length - 1] = s;
					let c = o.substring(s.raw.length).replace(/^\n/, ""), l = c ? c.split("\n").length : 0, d = l ? e.slice(0, -l) : e;
					if (d.length > 0 && (n = `${n}
${d.join("\n")}`), r = r.substring(0, r.length - t.text.length) + s.text, e = e.slice(d.length), e.length > 0 && this.rules.other.blockquoteStart.test(e[0])) continue;
					break;
				}
				if (u?.type === "list") {
					let t = u, a = t.raw + "\n" + e.join("\n"), o = this.list(a);
					i[i.length - 1] = o;
					let s = o.raw.length > t.raw.length ? a.slice(t.raw.length + 1, o.raw.length).split("\n").length : 0;
					if (s > 0) {
						let t = e.slice(0, s).join("\n");
						n = `${n}
${t}`, r = `${r}
${t}`;
					}
					e = e.slice(s);
				}
			}
			return {
				type: "blockquote",
				raw: n,
				tokens: i,
				text: r
			};
		}
	}
	list(e) {
		let t = this.rules.block.list.exec(e);
		if (t) {
			let n = e, r = () => n.length - e.length, i = t[1].trim(), a = i.length > 1, o = {
				type: "list",
				raw: "",
				ordered: a,
				start: a ? +i.slice(0, -1) : "",
				loose: !1,
				items: []
			};
			i = a ? `\\d{1,9}\\${i.slice(-1)}` : `\\${i}`, this.options.pedantic && (i = a ? i : "[*+-]");
			let s = this.rules.other.listItemRegex(i), c = !1;
			for (; e;) {
				let i = !1, a = "";
				if (!(t = s.exec(e)) || this.rules.block.hr.test(e)) break;
				let l = r();
				e = e.substring(t[0].length);
				let u = t[2].split("\n", 1)[0], d = t[1].length, f = this.options.pedantic ? za(u, d) : u.replace(this.rules.other.leadingSpaceTab, (e) => za(e, d)), p = e.split("\n", 1)[0], m = !f.trim(), h = 0;
				if (this.options.pedantic ? (h = 2, a = f.trimStart()) : m ? h = d + 1 : (h = f.search(this.rules.other.nonSpaceChar), h = h > 4 ? 1 : h, a = f.slice(h), h += d), m && this.rules.other.blankLine.test(p) && (e = e.substring(p.length + 1), i = !0), !i) {
					let t = this.rules.other.nextBulletRegex(h), n = this.rules.other.hrRegex(h), r = this.rules.other.fencesBeginRegex(h), i = this.options.pedantic ? this.rules.other.pedanticHeadingBeginRegex(h) : this.rules.other.headingBeginRegex(h), o = this.rules.other.htmlBeginRegex(h), s = this.rules.other.blockquoteBeginRegex(h);
					for (; e;) {
						let c = e.split("\n", 1)[0], l;
						if (p = c, this.options.pedantic ? (p = p.replace(this.rules.other.listReplaceNesting, "  "), l = p) : l = p.replace(this.rules.other.leadingSpaceTab, (e) => za(e)), r.test(p) || i.test(p) || o.test(p) || s.test(p) || t.test(p) || n.test(p)) break;
						if (l.search(this.rules.other.nonSpaceChar) >= h || !p.trim()) a += "\n" + l.slice(h);
						else {
							if (m || f.replace(this.rules.other.tabCharGlobal, "    ").search(this.rules.other.nonSpaceChar) >= 4 || r.test(f) || i.test(f) || o.test(f) || n.test(f)) break;
							let e = !this.options.pedantic && /^ {0,3}(?:=+|-+) *$/.test(p) && this.rules.block.lheading.test(`${f}
${p}`) ? `    ${l.replace(/^ {0,3}/, "")}` : p;
							a += "\n" + e;
						}
						m = !p.trim(), e = e.substring(c.length + 1), f = l.slice(h);
					}
				}
				let g = n.slice(l, r());
				o.loose || (c ? o.loose = !0 : this.rules.other.doubleBlankLine.test(g) && (c = !0)), o.items.push({
					type: "list_item",
					raw: g,
					task: !!this.options.gfm && this.rules.other.listIsTask.test(a),
					loose: !1,
					text: a,
					tokens: []
				});
			}
			let l = o.items.at(-1);
			if (l) l.raw = l.raw.trimEnd(), l.text = l.text.trimEnd();
			else return;
			o.raw = n.slice(0, r()).trimEnd();
			for (let e of o.items) if (this.lexer.state.top = !1, e.tokens = this.lexer.blockTokens(e.text, []), !o.loose) {
				let t = e.tokens.filter((e) => e.type === "space");
				o.loose = t.length > 0 && t.some((e) => this.rules.other.anyLine.test(e.raw));
			}
			for (let e of o.items) {
				let t = e.tokens[0];
				if (e.task && (t?.type === "text" || t?.type === "paragraph")) {
					e.text = e.text.replace(this.rules.other.listReplaceTask, ""), t.raw = t.raw.replace(this.rules.other.listReplaceTask, ""), t.text = t.text.replace(this.rules.other.listReplaceTask, "");
					for (let e = this.lexer.inlineQueue.length - 1; e >= 0; e--) if (this.rules.other.listIsTask.test(this.lexer.inlineQueue[e].src)) {
						this.lexer.inlineQueue[e].src = this.lexer.inlineQueue[e].src.replace(this.rules.other.listReplaceTask, "");
						break;
					}
					let n = this.rules.other.listTaskCheckbox.exec(e.raw);
					if (n) {
						let t = {
							type: "checkbox",
							raw: n[0] + " ",
							checked: n[0] !== "[ ]"
						};
						e.checked = t.checked, o.loose ? e.tokens[0] && ["paragraph", "text"].includes(e.tokens[0].type) && "tokens" in e.tokens[0] && e.tokens[0].tokens ? (e.tokens[0].raw = t.raw + e.tokens[0].raw, e.tokens[0].text = t.raw + e.tokens[0].text, e.tokens[0].tokens.unshift(t)) : e.tokens.unshift({
							type: "paragraph",
							raw: t.raw,
							text: t.raw,
							tokens: [t]
						}) : e.tokens.unshift(t);
					}
				} else e.task &&= !1;
			}
			if (o.loose) for (let e of o.items) {
				e.loose = !0;
				for (let t of e.tokens) t.type === "text" && (t.type = "paragraph");
			}
			return o;
		}
	}
	html(e) {
		let t = this.rules.block.html.exec(e);
		if (t) {
			let e = Aa(t[0]);
			return {
				type: "html",
				block: !0,
				raw: e,
				pre: t[1] === "pre" || t[1] === "script" || t[1] === "style",
				text: e
			};
		}
	}
	def(e) {
		let t = this.rules.block.def.exec(e);
		if (t) {
			if (!this.rules.other.startAngleBracket.test(t[2]) && Ma(t[2], "()") !== -1) return;
			let e = ja(t[1]), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: ka(t[0], "\n"),
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = Oa(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: ka(t[0], "\n"),
			header: [],
			align: [],
			rows: []
		};
		if (n.length === r.length) {
			for (let e of r) this.rules.other.tableAlignRight.test(e) ? a.align.push("right") : this.rules.other.tableAlignCenter.test(e) ? a.align.push("center") : this.rules.other.tableAlignLeft.test(e) ? a.align.push("left") : a.align.push(null);
			for (let e = 0; e < n.length; e++) a.header.push({
				text: n[e],
				tokens: this.lexer.inline(n[e]),
				header: !0,
				align: a.align[e]
			});
			for (let e of i) a.rows.push(Oa(e, a.header.length).map((e, t) => ({
				text: e,
				tokens: this.lexer.inline(e),
				header: !1,
				align: a.align[t]
			})));
			return a;
		}
	}
	lheading(e) {
		let t = this.rules.block.lheading.exec(e);
		if (t) {
			let e = t[1].trim();
			return {
				type: "heading",
				raw: ka(t[0], "\n"),
				depth: t[2].charAt(0) === "=" ? 1 : 2,
				text: e,
				tokens: this.lexer.inline(e)
			};
		}
	}
	paragraph(e) {
		let t = this.rules.block.paragraph.exec(e);
		if (t) {
			let e = t[1].charAt(t[1].length - 1) === "\n" ? t[1].slice(0, -1) : t[1];
			return {
				type: "paragraph",
				raw: t[0],
				text: e,
				tokens: this.lexer.inline(e)
			};
		}
	}
	text(e) {
		let t = this.rules.block.text.exec(e);
		if (t) return {
			type: "text",
			raw: t[0],
			text: t[0],
			tokens: this.lexer.inline(t[0])
		};
	}
	escape(e) {
		let t = this.rules.inline.escape.exec(e);
		if (t) return {
			type: "escape",
			raw: t[0],
			text: t[1]
		};
	}
	tag(e) {
		let t = this.rules.inline.tag.exec(e);
		if (t) return !this.lexer.state.inLink && this.rules.other.startATag.test(t[0]) ? this.lexer.state.inLink = !0 : this.lexer.state.inLink && this.rules.other.endATag.test(t[0]) && (this.lexer.state.inLink = !1), !this.lexer.state.inRawBlock && this.rules.other.startPreScriptTag.test(t[0]) ? this.lexer.state.inRawBlock = !0 : this.lexer.state.inRawBlock && this.rules.other.endPreScriptTag.test(t[0]) && (this.lexer.state.inRawBlock = !1), {
			type: "html",
			raw: t[0],
			inLink: this.lexer.state.inLink,
			inRawBlock: this.lexer.state.inRawBlock,
			block: !1,
			text: t[0]
		};
	}
	link(e) {
		if (this.lexer.state.linkParenPossible === !1) return;
		let t = this.rules.inline.link.exec(e);
		if (t) {
			let n = t[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && Ha(e, t[1], n, this.rules)) return;
			let r = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(r)) {
				if (!this.rules.other.endAngleBracket.test(r)) return;
				let e = ka(r.slice(0, -1), "\\");
				if ((r.length - e.length) % 2 == 0) return;
			} else {
				let e = Ma(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let i = ka(t[0].slice(0, -1), "\\");
			if ((t[0].length - i.length) % 2 == 0) return;
			let a = t[2], o = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(a);
				e && (a = e[1], o = e[3]);
			} else o = t[3] ? t[3].slice(1, -1) : "";
			return a = a.trim(), this.rules.other.startAngleBracket.test(a) && (a = this.options.pedantic && !this.rules.other.endAngleBracket.test(r) ? a.slice(1) : a.slice(1, -1)), Ba(t, {
				href: a && a.replace(this.rules.inline.anyPunctuation, "$1"),
				title: o && o.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let r = n[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && Ha(e, n[1], r, this.rules)) return;
			let i = t[ja(n[2] || n[1])];
			if (!i) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return Ba(n, i, n[0], this.lexer, this.rules);
		}
	}
	emStrong(e, t, n = "") {
		let r = this.rules.inline.emStrongLDelim.exec(e);
		if (!r || !r[1] && !r[2] && !r[3] && !r[4]) return;
		let i = r[0][0], a = n === i, o = t.length - e.length, s = o;
		for (; !this.options.pedantic && s > 0 && t[s - 1] === i;) s--;
		let c = [...t.slice(Math.max(0, s - 2), s)].at(-1) ?? "", l = c === "*" || c === "_" || !!this.rules.inline.punctuation.exec(c);
		s < o && (n = c);
		let u = s < o ? l : !!this.rules.inline.punctuation.exec(n);
		if (r[4] && n.match(this.rules.other.unicodeAlphaNumeric)) return;
		let d = r[1] || r[3] || "";
		if (!d || !n || u) {
			let n = [...r[0]].length - 1, u, f, p = n, m = 0, h = o - s + n, g = !this.options.pedantic && !!c.trim() && (!!d || i === "*" && !l);
			if (i !== "*" && i !== "_") return;
			let _ = i === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd, v = La(this, t, i), y = t.length - e.length + n;
			if (v < y) return;
			t = t.slice(-1 * e.length + n);
			let b = v - y;
			for (_.lastIndex = 0; (r = _.exec(t)) !== null && !(r.index > b);) {
				if (u = r[1] || r[2] || r[3] || r[4] || r[5] || r[6], !u) continue;
				if (f = [...u].length, r[3] || r[4]) {
					p += f;
					continue;
				}
				if (r[5] || r[6] || g) {
					if (a && (r[5] || r[6])) break;
					if (h % 3 && !((h + f) % 3)) {
						m += f;
						continue;
					}
				}
				if (p -= f, p > 0) continue;
				f = Math.min(f, f + p + m);
				let t = [...r[0]][0].length, i = e.slice(0, n + r.index + t + f);
				if (Math.min(n, f) % 2) {
					let e = i.slice(1, -1);
					return {
						type: "em",
						raw: i,
						text: e,
						tokens: this.lexer.inlineTokens(e)
					};
				}
				let o = i.slice(2, -2);
				return {
					type: "strong",
					raw: i,
					text: o,
					tokens: this.lexer.inlineTokens(o)
				};
			}
		}
	}
	codespan(e) {
		let t = this.rules.inline.code.exec(e);
		if (t) {
			let e = t[2].replace(this.rules.other.newLineCharGlobal, " "), n = this.rules.other.nonSpaceChar.test(e), r = this.rules.other.startingSpaceChar.test(e) && this.rules.other.endingSpaceChar.test(e);
			return n && r && (e = e.substring(1, e.length - 1)), {
				type: "codespan",
				raw: t[0],
				text: e
			};
		}
	}
	br(e) {
		let t = this.rules.inline.br.exec(e);
		if (t) return {
			type: "br",
			raw: t[0]
		};
	}
	del(e, t, n = "") {
		let r = this.rules.inline.delLDelim.exec(e);
		if (r && (!r[1] || !n || this.rules.inline.punctuation.exec(n))) {
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim, c = Ra(this, t), l = t.length - e.length + n;
			if (c < l) return;
			t = t.slice(-1 * e.length + n);
			let u = c - l;
			for (s.lastIndex = 0; (r = s.exec(t)) !== null && !(r.index > u);) {
				if (i = r[1] || r[2] || r[3] || r[4] || r[5] || r[6], !i || (a = [...i].length, a !== n)) continue;
				if (r[3] || r[4]) {
					o += a;
					continue;
				}
				if (o -= a, o > 0) continue;
				a = Math.min(a, a + o);
				let t = [...r[0]][0].length, s = e.slice(0, n + r.index + t + a), c = s.slice(n, -n);
				return {
					type: "del",
					raw: s,
					text: c,
					tokens: this.lexer.inlineTokens(c)
				};
			}
		}
	}
	autolink(e) {
		let t = this.rules.inline.autolink.exec(e);
		if (t) {
			let e, n;
			return t[2] === "@" ? (e = t[1], n = "mailto:" + e) : (e = t[1], n = e), {
				type: "link",
				raw: t[0],
				text: e,
				href: n,
				autolink: !0,
				tokens: [{
					type: "text",
					raw: e,
					text: e
				}]
			};
		}
	}
	url(e) {
		let t;
		if (t = this.rules.inline.url.exec(e)) {
			let e, n;
			if (t[2] === "@") e = t[0], n = "mailto:" + e;
			else {
				let r;
				do
					r = t[0], t[0] = this.rules.inline._backpedal.exec(t[0])?.[0] ?? "";
				while (r !== t[0]);
				e = t[0], n = t[1] === "www." ? "http://" + t[0] : t[0];
			}
			return {
				type: "link",
				raw: t[0],
				text: e,
				href: n,
				autolink: !0,
				tokens: [{
					type: "text",
					raw: e,
					text: e
				}]
			};
		}
	}
	inlineText(e) {
		let t = this.rules.inline.text.exec(e);
		if (t) {
			let e = this.lexer.state.inRawBlock;
			return {
				type: "text",
				raw: t[0],
				text: e ? t[0] : Ea(t[0]),
				escaped: e
			};
		}
	}
}, Wa = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || ti, this.options.tokenizer = this.options.tokenizer || new Ua(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			linkEmitted: !1,
			linkParenPossible: !0,
			top: !0
		};
		let t = {
			other: oi,
			block: xa.normal,
			inline: Sa.normal
		};
		this.options.pedantic ? (t.block = xa.pedantic, t.inline = Sa.pedantic) : this.options.gfm && (t.block = xa.gfm, t.inline = this.options.breaks ? Sa.breaks : Sa.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: xa,
			inline: Sa
		};
	}
	static lex(t, n) {
		return new e(n).lex(t);
	}
	static lexInline(t, n) {
		return new e(n).inlineTokens(t);
	}
	lex(e) {
		e = e.replace(oi.carriageReturn, "\n"), this.blockTokens(e, this.tokens);
		for (let e = 0; e < this.inlineQueue.length; e++) {
			let t = this.inlineQueue[e];
			this.state.inRawBlock = !1, this.inlineTokens(t.src, t.tokens);
		}
		return this.inlineQueue = [], this.tokens;
	}
	blockTokens(e, t = [], n = !1) {
		this.tokenizer.lexer = this, this.options.pedantic && (e = e.replace(oi.tabCharGlobal, "    ").replace(oi.spaceLine, ""));
		let r = 1 / 0;
		for (; e;) {
			if (e.length < r) r = e.length;
			else {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
			let i;
			if (this.options.extensions?.block?.some((n) => (i = n.call({ lexer: this }, e, t)) ? (e = e.substring(i.raw.length), t.push(i), !0) : !1)) continue;
			if (i = this.tokenizer.space(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				i.raw.length === 1 && n !== void 0 ? n.raw += "\n" : t.push(i);
				continue;
			}
			if (i = this.tokenizer.code(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				n?.type === "paragraph" || n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + i.raw, n.text += "\n" + i.text, this.inlineQueue.at(-1).src = n.text) : t.push(i);
				continue;
			}
			if ((i = this.tokenizer.fences(e)) || (i = this.tokenizer.heading(e)) || (i = this.tokenizer.hr(e)) || (i = this.tokenizer.blockquote(e)) || (i = this.tokenizer.list(e)) || (i = this.tokenizer.html(e))) {
				e = e.substring(i.raw.length), t.push(i);
				continue;
			}
			if (i = this.tokenizer.def(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				n?.type === "paragraph" || n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + i.raw, n.text += "\n" + i.raw, this.inlineQueue.at(-1).src = n.text) : (this.tokens.links[i.tag] || (this.tokens.links[i.tag] = {
					href: i.href,
					title: i.title
				}), t.push(i));
				continue;
			}
			if ((i = this.tokenizer.table(e)) || (i = this.tokenizer.lheading(e))) {
				e = e.substring(i.raw.length), t.push(i);
				continue;
			}
			let a = e;
			if (this.options.extensions?.startBlock) {
				let t = 1 / 0, n = e.slice(1), r;
				this.options.extensions.startBlock.forEach((e) => {
					r = e.call({ lexer: this }, n), typeof r == "number" && r >= 0 && (t = Math.min(t, r));
				}), t < 1 / 0 && t >= 0 && (a = e.substring(0, t + 1));
			}
			if (this.state.top && (i = this.tokenizer.paragraph(a))) {
				let r = t.at(-1);
				n && r?.type === "paragraph" ? (r.raw += (r.raw.endsWith("\n") ? "" : "\n") + i.raw, r.text += "\n" + i.text, this.inlineQueue.pop(), this.inlineQueue.at(-1).src = r.text) : t.push(i), n = a.length !== e.length, e = e.substring(i.raw.length);
			} else if (i = this.tokenizer.text(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + i.raw, n.text += "\n" + i.text, this.inlineQueue.pop(), this.inlineQueue.at(-1).src = n.text) : t.push(i);
			} else if (e) {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
		}
		return this.state.top = !0, t;
	}
	inline(e, t = []) {
		return this.inlineQueue.push({
			src: e,
			tokens: t
		}), t;
	}
	linkInText(e) {
		if (!e.includes("[")) return !1;
		let t = this.tokenizer.rules.inline.link;
		for (let n of e.matchAll(this.tokenizer.rules.inline.blockSkip)) if (t.test(n[0]) && e.charAt(n.index - 1) !== "!") return !0;
		for (let t of e.matchAll(this.tokenizer.rules.inline.reflinkSearch)) {
			let e = t[0], n = e.lastIndexOf("[");
			if (e.charAt(0) !== "!" && Object.hasOwn(this.tokens.links, ja(e.slice(n + 1, -1))) && !(n > 1 && this.linkInText(e.slice(1, n - 1)))) return !0;
		}
		return !1;
	}
	inlineTokens(e, t = []) {
		this.tokenizer.lexer = this;
		let n = this.state.linkParenPossible;
		this.state.linkParenPossible = n && e.includes(")");
		try {
			return this.#e(e, t);
		} finally {
			this.state.linkParenPossible = n;
		}
	}
	#e(e, t) {
		let n = e.includes("*") || e.includes("_") || e.includes("~") || this.options.hooks?.emStrongMask || this.tokenizer.emStrong !== Ua.prototype.emStrong || this.tokenizer.del !== Ua.prototype.del ? this.#n(e) : e;
		Fa(this.tokenizer, n);
		try {
			return this.#t(e, n, t);
		} finally {
			Ia(this.tokenizer);
		}
	}
	#t(e, t, n) {
		let r = !1, i = "", a = 1 / 0;
		for (; e;) {
			if (e.length < a) a = e.length;
			else {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
			r || (i = ""), r = !1;
			let o;
			if (this.options.extensions?.inline?.some((t) => (o = t.call({ lexer: this }, e, n)) ? (e = e.substring(o.raw.length), n.push(o), !0) : !1)) continue;
			if ((o = this.tokenizer.escape(e)) || (o = this.tokenizer.tag(e)) || (o = this.tokenizer.link(e))) {
				e = e.substring(o.raw.length), n.push(o);
				continue;
			}
			if (o = this.tokenizer.reflink(e, this.tokens.links)) {
				e = e.substring(o.raw.length);
				let t = n.at(-1);
				o.type === "text" && t?.type === "text" ? (t.raw += o.raw, t.text += o.text) : n.push(o);
				continue;
			}
			if ((o = this.tokenizer.emStrong(e, t, i)) || (o = this.tokenizer.codespan(e)) || (o = this.tokenizer.br(e)) || (o = this.tokenizer.del(e, t, i)) || (o = this.tokenizer.autolink(e)) || !this.state.inLink && (o = this.tokenizer.url(e))) {
				e = e.substring(o.raw.length), n.push(o);
				continue;
			}
			let s = e;
			if (this.options.extensions?.startInline) {
				let t = 1 / 0, n = e.slice(1), r;
				this.options.extensions.startInline.forEach((e) => {
					r = e.call({ lexer: this }, n), typeof r == "number" && r >= 0 && (t = Math.min(t, r));
				}), t < 1 / 0 && t >= 0 && (s = e.substring(0, t + 1));
			}
			if (o = this.tokenizer.inlineText(s)) {
				e = e.substring(o.raw.length), o.raw.slice(-1) !== "_" && (i = o.raw.slice(-1)), r = !0;
				let t = n.at(-1);
				t?.type === "text" ? (t.raw += o.raw, t.text += o.text) : n.push(o);
			} else if (e) {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
		}
		return n;
	}
	#n(e) {
		let t = e;
		if (this.tokens.links && e.includes("[")) {
			let e = this.tokenizer.rules.inline.reflinkSearch, n = (t) => {
				let r = t.lastIndexOf("[");
				if (!Object.hasOwn(this.tokens.links, ja(t.slice(r + 1, -1)))) return t;
				if (r > 1 && t.charAt(0) !== "!") {
					let i = t.slice(1, r - 1);
					if (this.linkInText(i)) return "[" + i.replace(e, n) + "][" + "a".repeat(t.length - r - 2) + "]";
				}
				return "[" + "a".repeat(t.length - 2) + "]";
			};
			t = t.replace(e, n);
		}
		return t = t.replace(this.tokenizer.rules.inline.anyPunctuationOutsideCode, (e, t, n) => n ? "+".repeat(e.length) : t ? t + e.slice(t.length, -t.length).replace(/`/g, "+") + t : e), t = t.replace(this.tokenizer.rules.inline.blockSkip, (e, t, n) => {
			let r = n ? n.length : 0;
			return e.slice(0, r) + "[" + "a".repeat(e.length - r - 2) + "]";
		}), t = this.options.hooks?.emStrongMask?.call({ lexer: this }, t) ?? t, t;
	}
	infiniteLoopError(e) {
		let t = "Infinite loop on byte: " + e;
		if (this.options.silent) console.error(t);
		else throw Error(t);
	}
}, Ga = class {
	options;
	parser;
	constructor(e) {
		this.options = e || ti;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(oi.notSpaceStart)?.[0], i = e ? e.replace(oi.endingNewline, "") + "\n" : "";
		return r ? "<pre><code class=\"language-" + Ta(r) + "\">" + (n ? i : Ta(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : Ta(i, !0)) + "</code></pre>\n";
	}
	blockquote({ tokens: e }) {
		return `<blockquote>
${this.parser.parse(e)}</blockquote>
`;
	}
	html({ text: e }) {
		return e;
	}
	def(e) {
		return "";
	}
	heading({ tokens: e, depth: t }) {
		return `<h${t}>${this.parser.parseInline(e)}</h${t}>
`;
	}
	hr(e) {
		return "<hr>\n";
	}
	list(e) {
		let t = e.ordered, n = e.start, r = "";
		for (let t = 0; t < e.items.length; t++) {
			let n = e.items[t];
			r += this.listitem(n);
		}
		let i = t ? "ol" : "ul", a = t && n !== 1 ? " start=\"" + n + "\"" : "";
		return "<" + i + a + ">\n" + r + "</" + i + ">\n";
	}
	listitem(e) {
		return `<li>${this.parser.parse(e.tokens)}</li>
`;
	}
	checkbox({ checked: e }) {
		return "<input " + (e ? "checked=\"\" " : "") + "disabled=\"\" type=\"checkbox\"> ";
	}
	paragraph({ tokens: e }) {
		return `<p>${this.parser.parseInline(e)}</p>
`;
	}
	table(e) {
		let t = "", n = "";
		for (let t = 0; t < e.header.length; t++) n += this.tablecell(e.header[t]);
		t += this.tablerow({ text: n });
		let r = "";
		for (let t = 0; t < e.rows.length; t++) {
			let i = e.rows[t];
			n = "";
			for (let e = 0; e < i.length; e++) n += this.tablecell(i[e]);
			r += this.tablerow({ text: n });
		}
		return r &&= `<tbody>${r}</tbody>`, "<table>\n<thead>\n" + t + "</thead>\n" + r + "</table>\n";
	}
	tablerow({ text: e }) {
		return `<tr>
${e}</tr>
`;
	}
	tablecell(e) {
		let t = this.parser.parseInline(e.tokens), n = e.header ? "th" : "td";
		return (e.align ? `<${n} align="${e.align}">` : `<${n}>`) + t + `</${n}>
`;
	}
	strong({ tokens: e }) {
		return `<strong>${this.parser.parseInline(e)}</strong>`;
	}
	em({ tokens: e }) {
		return `<em>${this.parser.parseInline(e)}</em>`;
	}
	codespan({ text: e }) {
		return `<code>${Ta(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, text: n, tokens: r, autolink: i }) {
		let a = i ? Ta(n, !0) : this.parser.parseInline(r), o = Da(e);
		if (o === null) return a;
		e = Ta(o, i);
		let s = "<a href=\"" + e + "\"";
		return t && (s += " title=\"" + Ta(t) + "\""), s += ">" + a + "</a>", s;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = Da(e);
		if (i === null) return Ta(n);
		e = i;
		let a = `<img src="${Ta(e)}" alt="${Ta(n)}"`;
		return t && (a += ` title="${Ta(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : Ta(e.text);
	}
}, Ka = class {
	parser;
	strong({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	em({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	codespan({ text: e }) {
		return e;
	}
	del({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	html({ text: e }) {
		return e;
	}
	text({ text: e }) {
		return e;
	}
	link({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	image({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	br() {
		return "";
	}
	checkbox({ raw: e }) {
		return e;
	}
}, qa = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || ti, this.options.renderer = this.options.renderer || new Ga(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new Ka(), this.textRenderer.parser = this;
	}
	static parse(t, n) {
		return new e(n).parse(t);
	}
	static parseInline(t, n) {
		return new e(n).parseInline(t);
	}
	parse(e) {
		this.renderer.parser = this;
		let t = "";
		for (let n = 0; n < e.length; n++) {
			let r = e[n];
			if (this.options.extensions?.renderers?.[r.type]) {
				let e = r, n = this.options.extensions.renderers[e.type].call({ parser: this }, e);
				if (n !== !1 || ![
					"space",
					"hr",
					"heading",
					"code",
					"table",
					"blockquote",
					"list",
					"checkbox",
					"html",
					"def",
					"paragraph",
					"text"
				].includes(e.type)) {
					t += n || "";
					continue;
				}
			}
			let i = r;
			switch (i.type) {
				case "space":
					t += this.renderer.space(i);
					break;
				case "hr":
					t += this.renderer.hr(i);
					break;
				case "heading":
					t += this.renderer.heading(i);
					break;
				case "code":
					t += this.renderer.code(i);
					break;
				case "table":
					t += this.renderer.table(i);
					break;
				case "blockquote":
					t += this.renderer.blockquote(i);
					break;
				case "list":
					t += this.renderer.list(i);
					break;
				case "checkbox":
					t += this.renderer.checkbox(i);
					break;
				case "html":
					t += this.renderer.html(i);
					break;
				case "def":
					t += this.renderer.def(i);
					break;
				case "paragraph":
					t += this.renderer.paragraph(i);
					break;
				case "text":
					t += this.renderer.text(i);
					break;
				default: {
					let e = "Token with \"" + i.type + "\" type was not found.";
					if (this.options.silent) return console.error(e), "";
					throw Error(e);
				}
			}
		}
		return t;
	}
	parseInline(e, t = this.renderer) {
		this.renderer.parser = this;
		let n = "";
		for (let r = 0; r < e.length; r++) {
			let i = e[r];
			if (this.options.extensions?.renderers?.[i.type]) {
				let e = this.options.extensions.renderers[i.type].call({ parser: this }, i);
				if (e !== !1 || ![
					"escape",
					"html",
					"link",
					"image",
					"checkbox",
					"strong",
					"em",
					"codespan",
					"br",
					"del",
					"text"
				].includes(i.type)) {
					n += e || "";
					continue;
				}
			}
			let a = i;
			switch (a.type) {
				case "escape":
					n += t.text(a);
					break;
				case "html":
					n += t.html(a);
					break;
				case "link":
					n += t.link(a);
					break;
				case "image":
					n += t.image(a);
					break;
				case "checkbox":
					n += t.checkbox(a);
					break;
				case "strong":
					n += t.strong(a);
					break;
				case "em":
					n += t.em(a);
					break;
				case "codespan":
					n += t.codespan(a);
					break;
				case "br":
					n += t.br(a);
					break;
				case "del":
					n += t.del(a);
					break;
				case "text":
					n += t.text(a);
					break;
				default: {
					let e = "Token with \"" + a.type + "\" type was not found.";
					if (this.options.silent) return console.error(e), "";
					throw Error(e);
				}
			}
		}
		return n;
	}
}, Ja = class {
	options;
	block;
	constructor(e) {
		this.options = e || ti;
	}
	static passThroughHooks = /* @__PURE__ */ new Set([
		"preprocess",
		"postprocess",
		"processAllTokens",
		"emStrongMask"
	]);
	static passThroughHooksRespectAsync = /* @__PURE__ */ new Set([
		"preprocess",
		"postprocess",
		"processAllTokens"
	]);
	preprocess(e) {
		return e;
	}
	postprocess(e) {
		return e;
	}
	processAllTokens(e) {
		return e;
	}
	emStrongMask(e) {
		return e;
	}
	provideLexer(e = this.block) {
		return e ? Wa.lex : Wa.lexInline;
	}
	provideParser(e = this.block) {
		return e ? qa.parse : qa.parseInline;
	}
};
function Ya(e, t) {
	if (Array.isArray(t)) for (let n = 0; n < t.length; n++) e.push(t[n]);
	else e.push(t);
}
var Xa = new class {
	defaults = ei();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = qa;
	Renderer = Ga;
	TextRenderer = Ka;
	Lexer = Wa;
	Tokenizer = Ua;
	Hooks = Ja;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (Ya(n, t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) Ya(n, this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) Ya(n, this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				Ya(n, this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					Ya(n, this.walkTokens(i, t));
				}) : e.tokens && Ya(n, this.walkTokens(e.tokens, t));
			}
		}
		return n;
	}
	use(...e) {
		let t = this.defaults.extensions || {
			renderers: Object.create(null),
			childTokens: Object.create(null)
		};
		return e.forEach((e) => {
			let n = { ...e };
			if (n.async = this.defaults.async || n.async || !1, e.extensions && (e.extensions.forEach((e) => {
				if (!e.name) throw Error("extension name required");
				if ("renderer" in e) {
					let n = t.renderers[e.name];
					n ? t.renderers[e.name] = function(...t) {
						let r = e.renderer.apply(this, t);
						return r === !1 && (r = n.apply(this, t)), r;
					} : t.renderers[e.name] = e.renderer;
				}
				if ("tokenizer" in e) {
					if (!e.level || e.level !== "block" && e.level !== "inline") throw Error("extension level must be 'block' or 'inline'");
					let n = t[e.level];
					n ? n.unshift(e.tokenizer) : t[e.level] = [e.tokenizer], e.start && (e.level === "block" ? t.startBlock ? t.startBlock.push(e.start) : t.startBlock = [e.start] : e.level === "inline" && (t.startInline ? t.startInline.push(e.start) : t.startInline = [e.start]));
				}
				"childTokens" in e && e.childTokens && (t.childTokens[e.name] = e.childTokens);
			}), n.extensions = t), e.renderer) {
				let t = this.defaults.renderer || new Ga(this.defaults);
				for (let n in e.renderer) {
					if (!(n in t)) throw Error(`renderer '${n}' does not exist`);
					if (["options", "parser"].includes(n)) continue;
					let r = n, i = e.renderer[r], a = t[r];
					t[r] = (...e) => {
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n === !1 ? "" : n ?? "";
					};
				}
				n.renderer = t;
			}
			if (e.tokenizer) {
				let t = this.defaults.tokenizer || new Ua(this.defaults);
				for (let n in e.tokenizer) {
					if (!(n in t)) throw Error(`tokenizer '${n}' does not exist`);
					if ([
						"options",
						"rules",
						"lexer"
					].includes(n)) continue;
					let r = n, i = e.tokenizer[r], a = t[r];
					t[r] = (...e) => {
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n;
					};
				}
				n.tokenizer = t;
			}
			if (e.hooks) {
				let t = this.defaults.hooks || new Ja();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = Ja.passThroughHooks.has(n) ? (e) => {
						if (t.options.async && Ja.passThroughHooksRespectAsync.has(n)) return (async () => {
							let n = await i.call(t, e);
							return a.call(t, n);
						})();
						let r = i.call(t, e);
						return a.call(t, r);
					} : (...e) => {
						if (t.options.async) return (async () => {
							let n = await i.apply(t, e);
							return n === !1 && (n = await a.apply(t, e)), n;
						})();
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n;
					};
				}
				n.hooks = t;
			}
			if (e.walkTokens) {
				let t = this.defaults.walkTokens, r = e.walkTokens;
				n.walkTokens = function(e) {
					let n = [];
					return Ya(n, r.call(this, e)), t && Ya(n, t.call(this, e)), n;
				};
			}
			this.defaults = {
				...this.defaults,
				...n
			};
		}), this;
	}
	setOptions(e) {
		return this.defaults = {
			...this.defaults,
			...e
		}, this;
	}
	lexer(e, t) {
		return Wa.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return qa.parse(e, t ?? this.defaults);
	}
	parseMarkdown(e) {
		return (t, n) => {
			let r = { ...n }, i = {
				...this.defaults,
				...r
			}, a = this.onError(!!i.silent, !!i.async);
			if (this.defaults.async === !0 && r.async === !1) return a(/* @__PURE__ */ Error("marked(): The async option was set to true by an extension. Remove async: false from the parse options object to return a Promise."));
			if (typeof t > "u" || t === null) return a(/* @__PURE__ */ Error("marked(): input parameter is undefined or null"));
			if (typeof t != "string") return a(/* @__PURE__ */ Error("marked(): input parameter is of type " + Object.prototype.toString.call(t) + ", string expected"));
			if (i.hooks && (i.hooks.options = i, i.hooks.block = e), i.async) return (async () => {
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? Wa.lex : Wa.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? qa.parse : qa.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? Wa.lex : Wa.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? qa.parse : qa.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + Ta(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function X(e, t) {
	return Xa.parse(e, t);
}
X.options = X.setOptions = function(e) {
	return Xa.setOptions(e), X.defaults = Xa.defaults, ni(X.defaults), X;
}, X.getDefaults = ei, X.defaults = ti;
function Za(...e) {
	return Xa.use(...e), X.defaults = Xa.defaults, ni(X.defaults), X;
}
X.use = Za, X.walkTokens = function(e, t) {
	return Xa.walkTokens(e, t);
}, X.parseInline = Xa.parseInline, X.Parser = qa, X.parser = qa.parse, X.Renderer = Ga, X.TextRenderer = Ka, X.Lexer = Wa, X.lexer = Wa.lex, X.Tokenizer = Ua, X.Hooks = Ja, X.parse = X, X.options, X.setOptions, X.walkTokens, X.parseInline, qa.parse, Wa.lex;
// node_modules/dompurify/dist/purify.es.mjs
function Qa(e, t) {
	this.v = e, this.k = t;
}
function $a(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function eo(e) {
	if (Array.isArray(e)) return e;
}
function to(e, t) {
	var n = e == null ? null : typeof Symbol < "u" && e[Symbol.iterator] || e["@@iterator"];
	if (n != null) {
		var r, i, a, o, s = [], c = !0, l = !1;
		try {
			if (a = (n = n.call(e)).next, t === 0) {
				if (Object(n) !== n) return;
				c = !1;
			} else for (; !(c = (r = a.call(n)).done) && (s.push(r.value), s.length !== t); c = !0);
		} catch (e) {
			l = !0, i = e;
		} finally {
			try {
				if (!c && n.return != null && (o = n.return(), Object(o) !== o)) return;
			} finally {
				if (l) throw i;
			}
		}
		return s;
	}
}
function no() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function ro(e, t) {
	return eo(e) || to(e, t) || io(e, t) || no();
}
function io(e, t) {
	if (e) {
		if (typeof e == "string") return $a(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? $a(e, t) : void 0;
	}
}
function ao(e) {
	var t, n;
	function r(t, n) {
		try {
			var a = e[t](n), o = a.value, s = o instanceof Qa;
			Promise.resolve(s ? o.v : o).then(function(n) {
				if (s) {
					var c = t === "return" && o.k ? t : "next";
					if (!o.k || n.done) return r(c, n);
					n = e[c](n).value;
				}
				i(!!a.done, n);
			}, function(e) {
				r("throw", e);
			});
		} catch (e) {
			i(2, e);
		}
	}
	function i(e, i) {
		e === 2 ? t.reject(i) : t.resolve({
			value: i,
			done: e
		}), (t = t.next) ? r(t.key, t.arg) : n = null;
	}
	this._invoke = function(e, i) {
		return new Promise(function(a, o) {
			var s = {
				key: e,
				arg: i,
				resolve: a,
				reject: o,
				next: null
			};
			n ? n = n.next = s : (t = n = s, r(e, i));
		});
	}, typeof e.return != "function" && (this.return = void 0);
}
ao.prototype[typeof Symbol == "function" && Symbol.asyncIterator || "@@asyncIterator"] = function() {
	return this;
}, ao.prototype.next = function(e) {
	return this._invoke("next", e);
}, ao.prototype.throw = function(e) {
	return this._invoke("throw", e);
}, ao.prototype.return = function(e) {
	return this._invoke("return", e);
};
var oo = Object.entries, so = Object.setPrototypeOf, co = Object.isFrozen, lo = Object.getPrototypeOf, uo = Object.getOwnPropertyDescriptor, fo = Object.freeze, po = Object.seal, mo = Object.create, ho = typeof Reflect < "u" && Reflect, go = ho.apply, _o = ho.construct;
fo ||= function(e) {
	return e;
}, po ||= function(e) {
	return e;
}, go ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, _o ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var vo = Z(Array.prototype.forEach);
Array.prototype.indexOf;
var yo = Z(Array.prototype.lastIndexOf), bo = Z(Array.prototype.pop), xo = Z(Array.prototype.push), So = Z(Array.prototype.slice), Co = Z(Array.prototype.splice), wo = Array.isArray, To = Ho(Set), Eo = Z(Set.prototype.add), Do = Z(Set.prototype.has), Oo = Z(String.prototype.toLowerCase), ko = Z(String.prototype.toString), Ao = Z(String.prototype.match), jo = Z(String.prototype.replace), Mo = Z(String.prototype.indexOf), No = Z(String.prototype.trim), Po = Z(Number.prototype.toString), Fo = Z(Boolean.prototype.toString), Io = typeof BigInt > "u" ? null : Z(BigInt.prototype.toString), Lo = typeof Symbol > "u" ? null : Z(Symbol.prototype.toString), Ro = Z(Object.prototype.hasOwnProperty), zo = Z(Object.prototype.toString), Bo = Z(RegExp.prototype.test), Vo = Ho(TypeError);
function Z(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return go(e, t, n);
	};
}
function Ho(e) {
	return function() {
		return _o(e, [...arguments]);
	};
}
function Q(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : Oo;
	if (so && so(e, null), !wo(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (co(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function Uo(e) {
	for (let t = 0; t < e.length; t++) Ro(e, t) || (e[t] = null);
	return e;
}
function Wo(e) {
	let t = mo(null);
	for (let r of oo(e)) {
		var n = ro(r, 2);
		let i = n[0], a = n[1];
		Ro(e, i) && (t[i] = wo(a) ? Uo(a) : a && typeof a == "object" && a.constructor === Object ? Wo(a) : a);
	}
	return t;
}
function Go(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return Po(e);
		case "boolean": return Fo(e);
		case "bigint": return Io ? Io(e) : "0";
		case "symbol": return Lo ? Lo(e) : "Symbol()";
		case "undefined": return zo(e);
		case "function":
		case "object": {
			if (e === null) return zo(e);
			let t = e, n = Ko(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : zo(e);
			}
			return zo(e);
		}
		default: return zo(e);
	}
}
function Ko(e, t) {
	for (; e !== null;) {
		let n = uo(e, t);
		if (n) {
			if (n.get) return Z(n.get);
			if (typeof n.value == "function") return Z(n.value);
		}
		e = lo(e);
	}
	function n() {
		return null;
	}
	return n;
}
function qo(e, t) {
	for (; e !== null;) {
		let n = uo(e, t);
		if (n) return n.get ? Z(n.get) : null;
		e = lo(e);
	}
	return null;
}
function Jo(e) {
	try {
		return Bo(e, ""), !0;
	} catch {
		return !1;
	}
}
var Yo = fo(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), Xo = fo(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), Zo = fo([
	"feBlend",
	"feColorMatrix",
	"feComponentTransfer",
	"feComposite",
	"feConvolveMatrix",
	"feDiffuseLighting",
	"feDisplacementMap",
	"feDistantLight",
	"feDropShadow",
	"feFlood",
	"feFuncA",
	"feFuncB",
	"feFuncG",
	"feFuncR",
	"feGaussianBlur",
	"feImage",
	"feMerge",
	"feMergeNode",
	"feMorphology",
	"feOffset",
	"fePointLight",
	"feSpecularLighting",
	"feSpotLight",
	"feTile",
	"feTurbulence"
]), Qo = fo([
	"animate",
	"color-profile",
	"cursor",
	"discard",
	"font-face",
	"font-face-format",
	"font-face-name",
	"font-face-src",
	"font-face-uri",
	"foreignobject",
	"hatch",
	"hatchpath",
	"mesh",
	"meshgradient",
	"meshpatch",
	"meshrow",
	"missing-glyph",
	"script",
	"set",
	"solidcolor",
	"unknown",
	"use"
]), $o = fo(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), es = fo([
	"maction",
	"maligngroup",
	"malignmark",
	"mlongdiv",
	"mscarries",
	"mscarry",
	"msgroup",
	"mstack",
	"msline",
	"msrow",
	"semantics",
	"annotation",
	"annotation-xml",
	"mprescripts",
	"none"
]), ts = fo(["#text"]), ns = fo(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), rs = fo(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dominant-baseline.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.pointer-events.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-orientation.text-rendering.textlength.type.u1.u2.unicode.values.vector-effect.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), is = fo(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), as = fo([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), os = po(/{{[\w\W]*|^[\w\W]*}}/g), ss = po(/<%[\w\W]*|^[\w\W]*%>/g), cs = po(/\${[\w\W]*/g), ls = po(/^data-[\-\w.\u00B7-\uFFFF]+$/), us = po(/^aria-[\-\w]+$/), ds = po(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), fs = po(/^(?:\w+script|data):/i), ps = po(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), ms = po(/^html$/i), hs = po(/^[a-z][.\w]*(-[.\w]+)+$/i), gs = po(/<[/\w!]/g), _s = po(/<[/\w]/g), vs = po(/<\/no(script|embed|frames)/i), ys = po(/\/>/i), $ = {
	element: 1,
	attribute: 2,
	text: 3,
	cdataSection: 4,
	entityReference: 5,
	entityNode: 6,
	processingInstruction: 7,
	comment: 8,
	document: 9,
	documentType: 10,
	documentFragment: 11,
	notation: 12
}, bs = fo(Q({}, [
	"style",
	"script",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"plaintext",
	"noscript"
])), xs = function() {
	return typeof window > "u" ? null : window;
}, Ss = function(e, t) {
	if (typeof e != "object" || typeof e.createPolicy != "function") return null;
	let n = null, r = "data-tt-policy-suffix";
	t && t.hasAttribute(r) && (n = t.getAttribute(r));
	let i = "dompurify" + (n ? "#" + n : "");
	try {
		return e.createPolicy(i, {
			createHTML(e) {
				return e;
			},
			createScriptURL(e) {
				return e;
			}
		});
	} catch {
		return console.warn("TrustedTypes policy " + i + " could not be created."), null;
	}
}, Cs = function() {
	return {
		afterSanitizeAttributes: [],
		afterSanitizeElements: [],
		afterSanitizeShadowDOM: [],
		beforeSanitizeAttributes: [],
		beforeSanitizeElements: [],
		beforeSanitizeShadowDOM: [],
		uponSanitizeAttribute: [],
		uponSanitizeElement: [],
		uponSanitizeShadowNode: []
	};
}, ws = function(e, t, n, r) {
	return Ro(e, t) && wo(e[t]) ? Q(r.base ? Wo(r.base) : {}, e[t], r.transform) : n;
}, Ts = function(e, t, n) {
	let r = Ro(e, t) ? e[t] : void 0;
	return r && typeof r == "object" ? Wo(r) : n();
};
function Es() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : xs(), t = (e) => Es(e);
	if (t.version = "3.4.17", t.removed = [], !e || !e.document || e.document.nodeType !== $.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = Ko(d, "cloneNode"), p = Ko(d, "remove"), m = Ko(d, "removeAttributeNode"), h = Ko(d, "nextSibling"), g = Ko(d, "childNodes"), _ = Ko(d, "parentNode"), v = Ko(d, "shadowRoot"), y = Ko(d, "attributes"), b = o && o.prototype ? Ko(o.prototype, "nodeType") : null, x = o && o.prototype ? Ko(o.prototype, "nodeName") : null, S = o && o.prototype ? Ko(o.prototype, "ownerDocument") : null, C = function(e) {
		return b ? b(e) : e.nodeType;
	}, ee = function(e) {
		return x ? x(e) : e.nodeName;
	}, te = qo(d, "namespaceURI"), ne = function(e) {
		return te ? te(e) : e.namespaceURI;
	};
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let w, re = "", ie, ae = !1, oe = 0, se = function() {
		if (oe > 0) throw Vo("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
	}, ce = function(e) {
		se(), oe++;
		try {
			return w.createHTML(e);
		} finally {
			oe--;
		}
	}, le = function(e) {
		se(), oe++;
		try {
			return w.createScriptURL(e);
		} finally {
			oe--;
		}
	}, ue = function() {
		return ae ||= (ie = Ss(u, i), !0), ie;
	}, de = n, fe = de.implementation, pe = de.createNodeIterator, me = de.createDocumentFragment, he = de.getElementsByTagName, ge = r.importNode, T = Cs();
	t.isSupported = typeof oo == "function" && typeof _ == "function" && fe && fe.createHTMLDocument !== void 0;
	let _e = os, ve = ss, ye = cs, E = ls, be = us, D = fs, xe = ps, Se = hs, O = ds, k = null, Ce = Q({}, [
		...Yo,
		...Xo,
		...Zo,
		...$o,
		...ts
	]), A = null, we = Q({}, [
		...ns,
		...rs,
		...is,
		...as
	]), Te = Object.seal(mo(null, {
		tagNameCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		},
		attributeNameCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		},
		allowCustomizedBuiltInElements: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: !1
		}
	})), Ee = null, De = null, Oe = Object.seal(mo(null, {
		tagCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		},
		attributeCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		}
	})), ke = !0, Ae = !0, je = !1, Me = !0, Ne = !1, Pe = !0, Fe = !1, Ie = !1, Le = null, Re = null, ze = !1, Be = !1, Ve = !1, He = !1, Ue = !0, We = !1, Ge = "user-content-", Ke = !0, qe = !1, j = null, Je = {}, Ye = null, Xe = Q({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Ze = null, Qe = Q({}, [
		"audio",
		"video",
		"img",
		"source",
		"image",
		"track"
	]), $e = null, et = Q({}, [
		"alt",
		"class",
		"for",
		"id",
		"label",
		"name",
		"pattern",
		"placeholder",
		"role",
		"summary",
		"title",
		"value",
		"style",
		"xmlns"
	]), tt = "http://www.w3.org/1998/Math/MathML", nt = "http://www.w3.org/2000/svg", rt = "http://www.w3.org/1999/xhtml", it = rt, at = !1, ot = null, st = Q({}, [
		tt,
		nt,
		rt
	], ko), ct = fo([
		"mi",
		"mo",
		"mn",
		"ms",
		"mtext"
	]), lt = Q({}, ct), ut = fo(["annotation-xml"]), dt = Q({}, ut), M = Q({}, [
		"title",
		"style",
		"font",
		"a",
		"script"
	]), ft = null, pt = ["application/xhtml+xml", "text/html"], N = null, mt = null, ht = n.createElement("form"), gt = function(e) {
		return e instanceof RegExp || e instanceof Function;
	}, _t = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		if (mt && mt === e) return;
		e && typeof e == "object" || (e = {}), e = Wo(e), ft = pt.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, N = ft === "application/xhtml+xml" ? ko : Oo, k = ws(e, "ALLOWED_TAGS", Ce, { transform: N }), A = ws(e, "ALLOWED_ATTR", we, { transform: N }), ot = ws(e, "ALLOWED_NAMESPACES", st, { transform: ko }), $e = ws(e, "ADD_URI_SAFE_ATTR", et, {
			transform: N,
			base: et
		}), Ze = ws(e, "ADD_DATA_URI_TAGS", Qe, {
			transform: N,
			base: Qe
		}), Ye = ws(e, "FORBID_CONTENTS", Xe, { transform: N }), Ee = ws(e, "FORBID_TAGS", Wo({}), { transform: N }), De = ws(e, "FORBID_ATTR", Wo({}), { transform: N }), Je = Ro(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? Wo(e.USE_PROFILES) : e.USE_PROFILES : !1, ke = e.ALLOW_ARIA_ATTR !== !1, Ae = e.ALLOW_DATA_ATTR !== !1, je = e.ALLOW_UNKNOWN_PROTOCOLS || !1, Me = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, Ne = e.SAFE_FOR_TEMPLATES || !1, Pe = e.SAFE_FOR_XML !== !1, Fe = e.WHOLE_DOCUMENT || !1, Be = e.RETURN_DOM || !1, Ve = e.RETURN_DOM_FRAGMENT || !1, He = e.RETURN_TRUSTED_TYPE || !1, ze = e.FORCE_BODY || !1, Ue = e.SANITIZE_DOM !== !1, We = e.SANITIZE_NAMED_PROPS || !1, Ke = e.KEEP_CONTENT !== !1, qe = e.IN_PLACE || !1, O = Jo(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : ds, it = typeof e.NAMESPACE == "string" ? e.NAMESPACE : rt, lt = Ts(e, "MATHML_TEXT_INTEGRATION_POINTS", () => Q({}, ct)), dt = Ts(e, "HTML_INTEGRATION_POINTS", () => Q({}, ut));
		let t = Ts(e, "CUSTOM_ELEMENT_HANDLING", () => mo(null));
		if (Te = mo(null), Ro(t, "tagNameCheck") && gt(t.tagNameCheck) && (Te.tagNameCheck = t.tagNameCheck), Ro(t, "attributeNameCheck") && gt(t.attributeNameCheck) && (Te.attributeNameCheck = t.attributeNameCheck), Ro(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (Te.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), po(Te), Ne && (Ae = !1), Ve && (Be = !0), Je && (k = Q({}, ts), A = mo(null), Je.html === !0 && (Q(k, Yo), Q(A, ns)), Je.svg === !0 && (Q(k, Xo), Q(A, rs), Q(A, as)), Je.svgFilters === !0 && (Q(k, Zo), Q(A, rs), Q(A, as)), Je.mathMl === !0 && (Q(k, $o), Q(A, is), Q(A, as))), Oe.tagCheck = null, Oe.attributeCheck = null, Ro(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? Oe.tagCheck = e.ADD_TAGS : wo(e.ADD_TAGS) && (k === Ce && (k = Wo(k)), Q(k, e.ADD_TAGS, N))), Ro(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? Oe.attributeCheck = e.ADD_ATTR : wo(e.ADD_ATTR) && (A === we && (A = Wo(A)), Q(A, e.ADD_ATTR, N))), Ro(e, "ADD_FORBID_CONTENTS") && wo(e.ADD_FORBID_CONTENTS) && (Ye === Xe && (Ye = Wo(Ye)), Q(Ye, e.ADD_FORBID_CONTENTS, N)), Ke && (k["#text"] = !0), Fe && Q(k, [
			"html",
			"head",
			"body"
		]), k.table && (Q(k, ["tbody"]), delete Ee.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw Vo("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw Vo("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
			let t = w;
			w = e.TRUSTED_TYPES_POLICY;
			try {
				re = ce("");
			} catch (e) {
				throw w = t, e;
			}
		} else e.TRUSTED_TYPES_POLICY === null ? (w = void 0, re = "") : (w === void 0 && (w = ue()), w && typeof re == "string" && (re = ce("")));
		fo && fo(e), mt = e;
	}, vt = Q({}, [
		...Xo,
		...Zo,
		...Qo
	]), yt = Q({}, [...$o, ...es]), bt = function(e, t, n) {
		return t.namespaceURI === rt ? e === "svg" : t.namespaceURI === tt ? e === "svg" && (n === "annotation-xml" || lt[n]) : !!vt[e];
	}, xt = function(e, t, n) {
		return t.namespaceURI === rt ? e === "math" : t.namespaceURI === nt ? e === "math" && dt[n] : !!yt[e];
	}, St = function(e, t, n) {
		return t.namespaceURI === nt && !dt[n] || t.namespaceURI === tt && !lt[n] ? !1 : !yt[e] && (M[e] || !vt[e]);
	}, Ct = function(e) {
		let t = ne(e), n = Oo(ee(e)), r = _(e), i, a;
		return r && C(r) === $.element ? (i = { namespaceURI: ne(r) }, a = Oo(ee(r))) : (i = { namespaceURI: it }, a = "template"), ot[t] ? t === nt ? bt(n, i, a) : t === tt ? xt(n, i, a) : t === rt ? St(n, i, a) : !!(ft === "application/xhtml+xml" && ot[t]) : !1;
	}, wt = function(e) {
		xo(t.removed, { element: e });
		try {
			_(e).removeChild(e);
		} catch {
			if (p(e), !_(e)) throw Vo("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
		}
	}, Tt = function(e, t, n) {
		try {
			m(e, t);
		} catch {
			try {
				e.removeAttribute(n);
			} catch {}
		}
	}, Et = function(e) {
		At(e);
		let t = g(e);
		t && vo(So(t), (e) => {
			try {
				p(e);
			} catch {}
		});
		let n = C(e) === $.element ? y(e) : null;
		if (n) for (let t = n.length - 1; t >= 0; --t) {
			let r = n[t], i = r && r.name;
			typeof i == "string" && Tt(e, r, i);
		}
	}, Dt = function(e, n, r) {
		if (!r) try {
			r = n.getAttributeNode(e);
		} catch {
			r = null;
		}
		xo(t.removed, {
			attribute: r || null,
			from: n
		});
		try {
			r ? m(n, r) : n.removeAttribute(e);
		} catch {
			try {
				n.removeAttribute(e);
			} catch {}
		}
		if (e === "is") {
			if (Be || Ve) try {
				wt(n);
			} catch {}
			else try {
				n.setAttribute(e, "");
			} catch {}
		}
	}, Ot = function(e) {
		let t = y(e);
		if (t) for (let n = t.length - 1; n >= 0; --n) {
			let r = t[n], i = r && r.name;
			typeof i != "string" || A[N(i)] || Tt(e, r, i);
		}
	}, kt = function(e, t) {
		let n = arguments.length > 2 && arguments[2] !== void 0 && arguments[2], r = [e];
		for (; r.length > 0;) {
			let e = r.pop();
			if (t(e) === !1) continue;
			if (n && C(e) === $.element) {
				let t = v(e);
				Ut(t) && r.push(t);
			}
			let i = g(e);
			if (i) for (let e = i.length - 1; e >= 0; --e) r.push(i[e]);
		}
	}, At = function(e) {
		kt(e, (e) => {
			C(e) === $.element && Ot(e);
		});
	}, P = function(e) {
		let t = [];
		return kt(e, (e) => {
			t.push(e);
		}, !0), t;
	}, jt = function(e, t) {
		let n = To();
		vo(P(e), (e) => {
			C(e) === $.element && Eo(n, e);
		}), vo(t, (e) => {
			C(e) === $.element && !Do(n, e) && Ot(e);
		});
	}, F = function(e, n) {
		let r = !1;
		return vo(t.removed, (t) => {
			t.element && (r ||= t.element === e, At(t.element));
		}), jt(e, n), r;
	}, Mt = function(e, t) {
		let n = j, r = null, i = () => (r || (r = To(), vo(t, (e) => {
			Eo(r, e);
		})), r);
		for (let t = 0;; t++) {
			let r = P(e), a = [e], o = !1;
			if (vo(r, (e) => {
				C(e) === $.documentFragment ? xo(a, e) : !o && !Do(n, e) && (o = Do(i(), e));
			}), !o) return;
			if (t >= 8) throw Vo("the tree kept changing during in-place sanitization; refusing to return it");
			vo(a, (t, n) => {
				Nt(t, n > 0, e, i());
			});
		}
	}, Nt = function(e, t, n, r) {
		let i = Rt(e), a;
		for (; a = i.nextNode();) Do(r, a) && !Do(j, a) && (t && L(T.uponSanitizeShadowNode, a, null), Kt(a, n), $t(a, n), Ut(a.content) && en(a.content));
	}, Pt = function(e) {
		j && Eo(j, e);
	}, Ft = function(e, t) {
		return Pe ? e === "patchsrc" || e === "for" && t !== "label" && t !== "output" : !1;
	}, It = function(e) {
		Pe && kt(e, (e) => {
			let t = C(e);
			if (t === $.processingInstruction || t === $.comment && Bo(_s, e.data)) {
				try {
					p(e);
				} catch {}
				return !1;
			}
			if (t === $.element) {
				let t = e, n = N(ee(e));
				try {
					t.hasAttribute && t.hasAttribute("patchsrc") && t.removeAttribute("patchsrc"), t.hasAttribute && t.hasAttribute("for") && Ft("for", n) && t.removeAttribute("for");
				} catch {}
			}
			return !0;
		});
	}, Lt = function(e) {
		let t = null, r = null;
		if (ze) e = "<remove></remove>" + e;
		else {
			let t = Ao(e, /^[\r\n\t ]+/);
			r = t && t[0];
		}
		ft === "application/xhtml+xml" && it === rt && (e = "<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>" + e + "</body></html>");
		let i = w ? ce(e) : e;
		if (it === rt) try {
			t = new l().parseFromString(i, ft);
		} catch {}
		if (!t || !t.documentElement) {
			t = fe.createDocument(it, "template", null);
			try {
				t.documentElement.innerHTML = at ? re : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), it === rt ? he.call(t, Fe ? "html" : "body")[0] : Fe ? t.documentElement : a;
	}, Rt = function(e) {
		let t = S ? S(e) : e.ownerDocument;
		return pe.call(t || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, zt = function(e) {
		return e = jo(e, _e, " "), e = jo(e, ve, " "), e = jo(e, ye, " "), e;
	}, Bt = function(e) {
		Vt(e);
		let t = e.querySelectorAll?.call(e, "template");
		t && vo(t, (e) => {
			Ut(e.content) && Bt(e.content);
		});
	}, Vt = function(e) {
		e.normalize();
		let t = S ? S(e) : e.ownerDocument, n = pe.call(t || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), r = n.nextNode();
		for (; r;) r.data = zt(r.data), r = n.nextNode();
	}, Ht = function(e) {
		let t = x ? x(e) : null;
		return typeof t != "string" || N(t) !== "form" ? !1 : typeof e.nodeName != "string" || typeof e.textContent != "string" || typeof e.removeChild != "function" || e.attributes !== y(e) || typeof e.removeAttribute != "function" || typeof e.removeAttributeNode != "function" || typeof e.getAttributeNode != "function" || typeof e.setAttribute != "function" || typeof e.namespaceURI != "string" || typeof e.insertBefore != "function" || typeof e.hasChildNodes != "function" || e.nodeType !== b(e) || e.childNodes !== g(e);
	}, Ut = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return b(e) === $.documentFragment;
		} catch {
			return !1;
		}
	}, I = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return typeof b(e) == "number";
		} catch {
			return !1;
		}
	};
	function L(e, n, r) {
		e.length !== 0 && vo(e, (e) => {
			e.call(t, n, r, mt);
		});
	}
	let Wt = function(e, t) {
		return !!(Pe && e.hasChildNodes() && !I(e.firstElementChild) && Bo(gs, e.textContent) && Bo(gs, e.innerHTML) || Pe && bs[t] && C(e) === $.element && ne(e) === rt && (I(e.firstElementChild) || typeof e.textContent == "string" && Bo(gs, e.textContent)) || e.nodeType === $.processingInstruction || Pe && e.nodeType === $.comment && Bo(_s, e.data));
	}, R = function(e, t) {
		return e instanceof RegExp ? Bo(e, t) : e instanceof Function && !!e(t, ...[...arguments].slice(2));
	}, Gt = function(e, t, n) {
		if (!Ee[t] && Yt(t) && R(Te.tagNameCheck, t)) return !1;
		if (Ke && !Ye[t] && !Ye[Oo(t)]) {
			let t = _(e), r = g(e);
			if (r && t) {
				let i = r.length;
				for (let a = i - 1; a >= 0; --a) {
					let i = e === n ? f(r[a], !0) : r[a];
					t.insertBefore(i, h(e));
				}
			}
		}
		return wt(e), !0;
	}, z = function(e, t, n, r) {
		return e.length === 0 ? t : t === n || t === r ? Wo(t) : t;
	}, B = function(e, t) {
		return e === t || _(e) !== null ? !1 : (qe && At(e), !0);
	}, Kt = function(e, n) {
		if (Pt(e), L(T.beforeSanitizeElements, e, null), B(e, n)) return !0;
		if (Ht(e)) return wt(e), !0;
		let r = N(ee(e));
		if (k = z(T.uponSanitizeElement, k, Ce, Le), L(T.uponSanitizeElement, e, {
			tagName: r,
			allowedTags: k
		}), B(e, n)) return !0;
		if (Wt(e, r)) return wt(e), !0;
		if (Ee[r] || !(Oe.tagCheck instanceof Function && Oe.tagCheck(r)) && !k[r]) {
			let t = Gt(e, r, n);
			return t === !1 && (L(T.afterSanitizeElements, e, null), B(e, n)) ? !0 : t;
		}
		if (C(e) === $.element && !Ct(e) || (r === "noscript" || r === "noembed" || r === "noframes") && Bo(vs, e.innerHTML)) return wt(e), !0;
		if (Ne && e.nodeType === $.text) {
			let n = zt(e.textContent);
			e.textContent !== n && (xo(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return L(T.afterSanitizeElements, e, null), B(e, n);
	}, qt = function(e, t, r) {
		if (De[t] || Ft(t, e) || Ue && (t === "id" || t === "name") && (r in n || r in ht)) return !1;
		let i = A[t] || Oe.attributeCheck instanceof Function && Oe.attributeCheck(t, e);
		return Ae && Bo(E, t) || ke && Bo(be, t) ? !0 : i ? $e[t] || Bo(O, jo(r, xe, "")) || (t === "src" || t === "xlink:href" || t === "href") && e !== "script" && Mo(r, "data:") === 0 && Ze[e] || je && !Bo(D, jo(r, xe, "")) ? !0 : !r : Yt(e) && R(Te.tagNameCheck, e) && R(Te.attributeNameCheck, t, e) || t === "is" && Te.allowCustomizedBuiltInElements && R(Te.tagNameCheck, r);
	}, Jt = Q({}, [
		"annotation-xml",
		"color-profile",
		"font-face",
		"font-face-format",
		"font-face-name",
		"font-face-src",
		"font-face-uri",
		"missing-glyph"
	]), Yt = function(e) {
		return !Jt[Oo(e)] && Bo(Se, e);
	}, Xt = function(e, t, n, r) {
		if (w && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return ce(r);
			case "TrustedScriptURL": return le(r);
		}
		return r;
	}, Zt = function(e, t, n, r) {
		try {
			return n ? e.setAttributeNS(n, t, r) : e.setAttribute(t, r), !Ht(e) || (wt(e), !1);
		} catch {
			return Dt(t, e), !1;
		}
	}, Qt = function(e, t) {
		return Pe && Bo(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, t) ? !0 : Oo(e) === "attributename" && Ao(t, "href") !== null;
	}, $t = function(e, n) {
		if (L(T.beforeSanitizeAttributes, e, null), B(e, n)) return;
		let r = e.attributes;
		if (!r || Ht(e)) return;
		A = z(T.uponSanitizeAttribute, A, we, Re);
		let i = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: A,
			forceKeepAttr: void 0
		}, a = r.length, o = N(e.nodeName);
		for (; a--;) {
			let n = r[a], s = n.name, c = n.namespaceURI, l = n.value, u = N(s), d = l, f = s === "value" ? d : No(d), p = !1;
			i.attrName = u, i.attrValue = f, i.keepAttr = !0, i.forceKeepAttr = void 0, L(T.uponSanitizeAttribute, e, i), f = i.attrValue, We && (u === "id" || u === "name") && Mo(f, Ge) !== 0 && (Dt(s, e, n), f = Ge + f, p = !0), Qt(s, f) ? Dt(s, e, n) : i.forceKeepAttr || (i.keepAttr ? !Me && Bo(ys, f) ? Dt(s, e, n) : (Ne && (f = zt(f)), qt(o, u, f) ? (f = Xt(o, u, c, f), f !== d && Zt(e, s, c, f) && p && bo(t.removed)) : Dt(s, e, n)) : Dt(s, e, n));
		}
		L(T.afterSanitizeAttributes, e, null), B(e, n);
	}, en = function(e) {
		let t = null, n = Rt(e);
		for (L(T.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (L(T.uponSanitizeShadowNode, t, null), Kt(t, e), $t(t, e), Ut(t.content) && en(t.content), C(t) === $.element) {
			let e = v(t);
			Ut(e) && (tn(e), en(e));
		}
		Ne && Vt(e), L(T.afterSanitizeShadowDOM, e, null);
	}, tn = function(e) {
		let t = [{
			node: e,
			shadow: null
		}];
		for (; t.length > 0;) {
			let e = t.pop();
			if (e.shadow) {
				en(e.shadow);
				continue;
			}
			let n = e.node, r = C(n) === $.element, i = g(n);
			if (i) for (let e = i.length - 1; e >= 0; --e) t.push({
				node: i[e],
				shadow: null
			});
			if (r) {
				let e = x ? x(n) : null;
				if (typeof e == "string" && N(e) === "template") {
					let e = n.content;
					Ut(e) && t.push({
						node: e,
						shadow: null
					});
				}
			}
			if (r) {
				let e = v(n);
				Ut(e) && t.push({
					node: null,
					shadow: e
				}, {
					node: e,
					shadow: null
				});
			}
		}
	};
	return t.sanitize = function(e) {
		let n = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {}, i = null, a = null, o = null, s = null;
		if (at = !e, at && (e = "<!-->"), typeof e != "string" && !I(e) && (e = Go(e), typeof e != "string")) throw Vo("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		Ie ? (k = Le, A = Re) : _t(n), (T.uponSanitizeElement.length > 0 || T.uponSanitizeAttribute.length > 0) && (k = Wo(k)), T.uponSanitizeAttribute.length > 0 && (A = Wo(A)), t.removed = [];
		let c = qe && typeof e != "string" && I(e), l = [], u = j;
		if (c) {
			It(e);
			let t = ee(e);
			if (typeof t == "string") {
				let n = N(t);
				if (!k[n] || Ee[n]) throw Et(e), Vo("root node is forbidden and cannot be sanitized in-place");
			}
			if (Ht(e)) throw Et(e), Vo("root node is clobbered and cannot be sanitized in-place");
			l = P(e), j = To();
		} else if (I(e)) i = Lt("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === $.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), tn(i);
		else {
			if (!Be && !Ne && !Fe && e.indexOf("<") === -1) return w && He ? ce(e) : e;
			if (i = Lt(e), !i) return Be ? null : He ? re : "";
		}
		i && ze && wt(i.firstChild);
		let d = c ? e : i;
		try {
			c && tn(e);
			let t = Rt(d);
			for (; o = t.nextNode();) Kt(o, d), $t(o, d), Ut(o.content) && en(o.content);
			c && Mt(e, l);
		} catch (t) {
			throw c && (Et(e), F(e, l)), t;
		} finally {
			j = u;
		}
		if (c) {
			if (F(e, l)) throw Vo("a node selected for removal could not be safely returned; refusing to sanitize in place");
			return Ne && Bt(e), e;
		}
		if (Be) {
			if (Ne && Bt(i), Ve) for (s = me.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (A.shadowroot || A.shadowrootmode) && (s = ge.call(r, s, !0)), s;
		}
		let f = Fe ? i.outerHTML : i.innerHTML;
		return Fe && k["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && Bo(ms, i.ownerDocument.doctype.name) && (f = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + f), Ne && (f = zt(f)), w && He ? ce(f) : f;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		_t(e), Ie = !0, Le = k, Re = A;
	}, t.clearConfig = function() {
		mt = null, Ie = !1, Le = null, Re = null, w = ie, re = "";
	}, t.isValidAttribute = function(e, t, n) {
		mt || _t({});
		let r = N(e), i = N(t);
		return typeof n == "string" && (Qt(t, n) || !Me && Bo(ys, n)) ? !1 : qt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && Ro(T, e) && xo(T[e], t);
	}, t.removeHook = function(e, t) {
		if (Ro(T, e)) {
			if (t !== void 0) {
				let n = yo(T[e], t);
				return n === -1 ? void 0 : Co(T[e], n, 1)[0];
			}
			return bo(T[e]);
		}
	}, t.removeHooks = function(e) {
		Ro(T, e) && (T[e] = []);
	}, t.removeAllHooks = function() {
		T = Cs();
	}, t;
}
var Ds = Es();
// kit/markdown.js
function Os(e) {
	let t = X.parse(e || "", {
		async: !1,
		gfm: !0,
		breaks: !1
	}), n = Ds.sanitize(t, {
		USE_PROFILES: { html: !0 },
		FORBID_TAGS: [
			"style",
			"form",
			"input",
			"button",
			"iframe",
			"object",
			"embed",
			"svg",
			"math",
			"img"
		],
		FORBID_ATTR: ["style"]
	}), r = new DOMParser().parseFromString(n, "text/html");
	for (let e of r.querySelectorAll("a")) {
		let t = !1;
		try {
			t = ["http:", "https:"].includes(new URL(e.href).protocol);
		} catch {}
		t || e.removeAttribute("href"), e.target = "_blank", e.rel = "noopener noreferrer";
	}
	return r.body.innerHTML;
}
// kit/Holder.svelte
var ks = /* @__PURE__ */ G("<span class=\"svelte-bwb1ob\"> </span>"), As = /* @__PURE__ */ G("<span class=\"name svelte-bwb1ob\" data-testid=\"holder-pool\"> </span>"), js = /* @__PURE__ */ G("<span class=\"name svelte-bwb1ob\" data-testid=\"holder-name\"> </span>"), Ms = /* @__PURE__ */ G("<span class=\"faint model svelte-bwb1ob\"> </span>"), Ns = /* @__PURE__ */ G("<span class=\"tool muted svelte-bwb1ob\" data-testid=\"holder-tool\"> </span>"), Ps = /* @__PURE__ */ G("<button type=\"button\" class=\"holder svelte-bwb1ob\" data-testid=\"holder\"><span></span> <!> <!> <!> <!></button>");
function Fs(e, t) {
	Be(t, !0);
	let n = Ur(t, "label", 3, ""), r = /* @__PURE__ */ it(() => t.h.display_name || t.h.id), i = /* @__PURE__ */ it(() => t.h.tool || t.h.tool_target ? $r(t.h.agent, {
		title: t.h.tool,
		target: t.h.tool_target,
		intent: t.h.tool_intent
	}).label : "working"), a = /* @__PURE__ */ it(() => [
		W(r),
		[t.h.pool, t.h.model].filter(Boolean).join(" · "),
		t.h.busy ? W(i) : "",
		t.h.id
	].filter(Boolean).join("\n") + "\nClick to open read-only");
	var o = Ps(), s = R(o);
	let c;
	var l = B(s, 2), u = (e) => {
		var t = ks(), r = z(t, !0);
		V(() => q(r, n())), K(e, t);
	};
	J(l, (e) => {
		n() && e(u);
	});
	var d = B(l, 2), f = (e) => {
		var n = As(), r = z(n, !0);
		V(() => q(r, t.h.pool)), K(e, n);
	}, p = (e) => {
		var t = js(), n = z(t, !0);
		V(() => q(n, W(r))), K(e, t);
	};
	J(d, (e) => {
		t.h.pool ? e(f) : e(p, -1);
	});
	var m = B(d, 2), h = (e) => {
		var n = Ms(), r = z(n, !0);
		V(() => q(r, t.h.model)), K(e, n);
	};
	J(m, (e) => {
		t.h.model && e(h);
	});
	var g = B(m, 2), _ = (e) => {
		var t = Ns(), n = z(t, !0);
		V(() => q(n, W(i))), K(e, t);
	};
	J(g, (e) => {
		t.h.busy && e(_);
	}), O(o), V(() => {
		Lr(o, "title", W(a)), Lr(o, "data-session", t.h.id), Lr(o, "data-name", W(r)), c = Nr(s, 1, "dot svelte-bwb1ob", null, c, {
			busy: t.h.busy,
			dead: t.h.alive === !1
		});
	}), Qn("click", o, (e) => {
		e.stopPropagation(), t.onopen(t.h.id);
	}), K(e, o), Ve();
}
$n(["click"]);
// plugins/attention/Attention.svelte
var Is = /* @__PURE__ */ G("<button class=\"ghost\" data-testid=\"push-test\">Test</button>"), Ls = /* @__PURE__ */ G("<!> <button data-testid=\"push-toggle\"> </button>", 1), Rs = /* @__PURE__ */ G("<span class=\"faint pushna svelte-t4vask\" data-testid=\"push-state\"> </span>"), zs = /* @__PURE__ */ G("<div class=\"card err\"> </div>"), Bs = /* @__PURE__ */ G("<p class=\"faint\">Loading…</p>"), Vs = /* @__PURE__ */ G("<div class=\"card empty\" data-testid=\"attention-empty\"><div class=\"big\">Nothing needs you.</div> <div class=\"muted\">No escalations, questions, decisions or other items waiting on you.</div></div>"), Hs = /* @__PURE__ */ G("<a target=\"_blank\" rel=\"noopener noreferrer\" data-testid=\"attention-ref\"> </a>"), Us = /* @__PURE__ */ G("· <!>", 1), Ws = /* @__PURE__ */ G("<div class=\"summary muted svelte-t4vask\" data-testid=\"attention-summary\"> </div>"), Gs = /* @__PURE__ */ G("<details class=\"qctx svelte-t4vask\" data-testid=\"attention-context\"><summary class=\"svelte-t4vask\">Context</summary> <div class=\"md svelte-t4vask\"></div></details>"), Ks = /* @__PURE__ */ G("<span class=\"muted\" data-testid=\"row-pending\"> </span>"), qs = /* @__PURE__ */ G("<span class=\"warn-text svelte-t4vask\" data-testid=\"row-stuck\"> </span>"), Js = /* @__PURE__ */ G("<span class=\"rec svelte-t4vask\"></span>"), Ys = /* @__PURE__ */ G("<span class=\"od svelte-t4vask\"> </span>"), Xs = /* @__PURE__ */ G("<button type=\"button\" data-testid=\"question-option\"><span class=\"ot svelte-t4vask\"> <!></span> <!></button>"), Zs = /* @__PURE__ */ G("<div class=\"opts svelte-t4vask\" data-testid=\"question-options\"></div>"), Qs = /* @__PURE__ */ G("<button type=\"button\" data-testid=\"row-resolve\" class=\"svelte-t4vask\"> </button>"), $s = /* @__PURE__ */ G("<button type=\"button\" class=\"ghost svelte-t4vask\" data-testid=\"open-chat\" title=\"Talk it over or answer in your own words\">Discuss</button>"), ec = /* @__PURE__ */ G("<button type=\"button\" class=\"ghost svelte-t4vask\" data-testid=\"open-chat\">Open chat</button>"), tc = /* @__PURE__ */ G("<!> <!> <!> <!>", 1), nc = /* @__PURE__ */ G("<tr data-testid=\"attention-row\"><td class=\"c-id svelte-t4vask\"><span class=\"pill warn\"> </span></td><td class=\"c-title svelte-t4vask\"> <div class=\"faint mono\"> <!></div> <!> <div class=\"acts svelte-t4vask\" role=\"presentation\"><!> <!></div></td><td class=\"muted full svelte-t4vask\"> </td><td class=\"muted c-time svelte-t4vask\"> </td></tr>"), rc = /* @__PURE__ */ G("<div class=\"card scroll\"><table class=\"stack\" data-testid=\"attention-list\"><thead><tr><th>Kind</th><th>Item</th><th>Why</th><th>Since</th></tr></thead><tbody></tbody></table></div>"), ic = /* @__PURE__ */ G("<span class=\"step muted svelte-t4vask\"> </span> <!>", 1), ac = /* @__PURE__ */ G("<span class=\"faint\" data-testid=\"no-holder\">waiting for a session</span>"), oc = /* @__PURE__ */ G("<a class=\"mono svelte-t4vask\"> </a>"), sc = /* @__PURE__ */ G("<span class=\"note svelte-t4vask\"> </span>"), cc = /* @__PURE__ */ G("<li class=\"svelte-t4vask\"><span> </span> <span class=\"step muted svelte-t4vask\"> </span> <!> <!> <span class=\"faint when svelte-t4vask\"> </span></li>"), lc = /* @__PURE__ */ G("<li class=\"svelte-t4vask\"><span class=\"pill\"> </span><span class=\"step muted svelte-t4vask\"> </span><span class=\"faint\">no report yet</span></li>"), uc = /* @__PURE__ */ G("<li class=\"svelte-t4vask\"><span class=\"faint\">waiting for a step to start</span></li>"), dc = /* @__PURE__ */ G("<a class=\"svelte-t4vask\"> </a>"), fc = /* @__PURE__ */ G("<li class=\"svelte-t4vask\"><span class=\"pill\">comment</span><span class=\"step muted mono svelte-t4vask\"><!></span><span class=\"note svelte-t4vask\"> </span><span class=\"faint when svelte-t4vask\"> </span></li>"), pc = /* @__PURE__ */ G("<div class=\"item card svelte-t4vask\" data-testid=\"progress-card\"><button class=\"head svelte-t4vask\"><span class=\"dot accent\"></span> <span class=\"title svelte-t4vask\"> </span> <span class=\"mono faint id svelte-t4vask\"> </span> <span class=\"faint when svelte-t4vask\"> </span></button> <div class=\"who svelte-t4vask\"><!></div> <ul class=\"svelte-t4vask\"><!> <!></ul></div>"), mc = /* @__PURE__ */ G("<div class=\"progress svelte-t4vask\" data-testid=\"progress\"></div>"), hc = /* @__PURE__ */ G("<div class=\"faint\">Loading…</div>"), gc = /* @__PURE__ */ G("<div class=\"faint none svelte-t4vask\" data-testid=\"progress-empty\">No runs in progress.</div>"), _c = /* @__PURE__ */ G("<header class=\"vh\"><h1>Attention</h1> <span class=\"sub\">What waits on a human, oldest first.</span> <span class=\"actions\"><!></span></header> <div class=\"pad\"><!> <!> <h2 class=\"st\">In progress</h2> <!></div>", 1);
function vc(e, t) {
	Be(t, !0);
	let n = /* @__PURE__ */ P(null), r = /* @__PURE__ */ P(""), i = /* @__PURE__ */ P(Rt([])), a = /* @__PURE__ */ P(Rt({})), o = /* @__PURE__ */ P(!1), s = /* @__PURE__ */ P(""), c = /* @__PURE__ */ P(Rt({
		state: "off",
		error: ""
	})), l = Rt({}), u = Rt({}), d = 0;
	async function f() {
		let e = Object.keys(l);
		try {
			let i = await t.api.get("api/attention"), a = new Set(i.map((e) => e.kind + ":" + e.id)), o = !1;
			for (let t of e) {
				let e = l[t];
				e && (a.has(t) ? ++e.loads >= 2 ? (u[t] = e.label, delete l[t], Xr(t)) : o = !0 : (delete l[t], Xr(t)));
			}
			F(n, i, !0), o && !d && (d = window.setTimeout(() => {
				d = 0, f();
			}, 1500)), F(r, "");
		} catch (e) {
			F(r, String(e.message || e), !0);
		}
		try {
			let e = await t.api.get("api/work?state=running&limit=6&detail=1"), n = e.detail ?? {};
			F(a, e.sessions ?? {}, !0), F(i, (e.runs ?? []).slice(0, 6).map((e) => ({
				run: e,
				steps: n[e.id]?.steps ?? [],
				comment: n[e.id]?.comment ?? null
			})), !0);
		} catch {}
		F(o, !0);
	}
	function p(e, t) {
		delete u[e], l[e] = {
			label: t,
			loads: 0
		};
	}
	rn(() => Jn(() => {
		for (let [e, t] of Jr()) p(e, t);
		let e = Zr((e, t) => {
			p(e, t), f();
		});
		f(), t.api.push.status().then((e) => F(c, e, !0));
		let n = 0, r = t.api.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, f();
			}, 1e3);
		}), i = t.api.params();
		if (i.item && !i.session) {
			let [e, ...t] = i.item.split(":");
			m({
				kind: e,
				id: t.join(":")
			});
		}
		return () => {
			r(), e(), clearTimeout(n), clearTimeout(d), d = 0;
		};
	}));
	async function m(e) {
		let n = e.kind + ":" + e.id;
		F(s, n);
		try {
			let r = await t.api.post("api/attention/chat", {
				kind: e.kind,
				id: e.id,
				title: e.title || "",
				why: e.why || ""
			});
			t.api.openSession(r.session, { item: n });
		} catch (r) {
			t.api.toast("Could not open a chat for " + n + ": " + String(r.message || r)), [
				"run",
				"ticket",
				"question",
				"decision",
				"stale",
				"wait"
			].includes(e.kind) && t.api.openTicket(e.id);
		} finally {
			F(s, "");
		}
	}
	let h = /* @__PURE__ */ P("");
	async function g(e, n) {
		let r = e.kind + ":" + e.id;
		F(h, r);
		try {
			await Gr(t.api, n) && Yr(r, n.label);
		} finally {
			F(h, "");
		}
	}
	let _ = (e, n) => ({
		label: "Answered " + n.title,
		run: () => t.api.call("question.answer", {
			id: e.id,
			option: n.id
		}),
		ask: "Answer " + e.id + " with option " + n.id + ": " + n.title + "?"
	}), v = {
		unsupported: "This browser has no Web Push (on iOS, add Desk to the home screen first).",
		unavailable: "Desk runs without --push.",
		denied: "Notifications are blocked in the browser settings.",
		off: "Get a notification for each new attention item.",
		on: "Notifications are on for this browser."
	};
	async function y() {
		F(c, W(c).state === "on" ? await t.api.push.unsubscribe() : await t.api.push.subscribe(), !0);
	}
	async function b() {
		F(c, await t.api.push.test(), !0), W(c).error || t.api.toast("Test notification sent");
	}
	function x(e) {
		if (!e || e.startsWith("0001")) return "";
		let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3);
		return t < 60 ? t + "s ago" : t < 3600 ? Math.round(t / 60) + "m ago" : t < 86400 ? Math.round(t / 3600) + "h ago" : Math.round(t / 86400) + "d ago";
	}
	let S = (e, t) => typeof e?.[t] == "string" ? JSON.parse(e[t] || "{}") : e?.[t] || {};
	function C(e) {
		return e.map((e) => ({
			s: e,
			r: S(e, "meta").report
		})).filter((e) => e.r && (e.r.note || e.r.state)).sort((e, t) => String(t.r.at).localeCompare(String(e.r.at)));
	}
	let ee = (e) => e === "failed" || e === "blocked" ? "bad" : e === "waiting" || e === "dropped" ? "warn" : e === "ready" ? "ok" : "accent", te = (e) => !!e && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(e), ne = (e) => e?.status === "in_progress" && e.holder ? W(a)[e.holder] ?? { id: e.holder } : null, w = (e) => t.api.openSession(e, { readonly: !0 });
	var re = _c(), ie = Gt(re), ae = B(R(ie), 4), oe = R(ae), se = (e) => {
		var t = Ls(), n = Gt(t), r = (e) => {
			var t = Is();
			Qn("click", t, b), K(e, t);
		};
		J(n, (e) => {
			W(c).state === "on" && e(r);
		});
		var i = B(n, 2);
		let a;
		var o = z(i, !0);
		V(() => {
			Lr(i, "title", v[W(c).state]), a = Nr(i, 1, "", null, a, { on: W(c).state === "on" }), q(o, W(c).state === "on" ? "Notifications on" : "Notify me");
		}), Qn("click", i, y), K(e, t);
	}, ce = (e) => {
		var t = Rs(), n = z(t, !0);
		V(() => {
			Lr(t, "title", v[W(c).state]), q(n, W(c).state === "unavailable" ? "" : "no notifications");
		}), K(e, t);
	};
	J(oe, (e) => {
		W(c).state === "on" || W(c).state === "off" ? e(se) : e(ce, -1);
	}), O(ae), O(ie);
	var le = B(ie, 2), ue = R(le), de = (e) => {
		var t = zs(), n = z(t, !0);
		V(() => q(n, W(c).error)), K(e, t);
	};
	J(ue, (e) => {
		W(c).error && e(de);
	});
	var fe = B(ue, 2), pe = (e) => {
		var t = zs(), n = z(t, !0);
		V(() => q(n, W(r))), K(e, t);
	}, me = (e) => {
		K(e, Bs());
	}, he = (e) => {
		K(e, Vs());
	}, ge = (e) => {
		var r = rc(), i = R(r), a = B(R(i));
		wr(a, 21, () => W(n), (e) => e.kind + e.id, (e, n) => {
			let r = /* @__PURE__ */ it(() => W(n).kind + ":" + W(n).id), i = /* @__PURE__ */ it(() => l[W(r)]);
			var a = nc();
			let o;
			var c = R(a), d = z(R(c), !0);
			O(c);
			var f = B(c), p = R(f, !0), v = B(p), y = R(v, !0), b = B(y), S = (e) => {
				var t = Us(), r = B(Gt(t)), i = (e) => {
					var t = Hs(), r = z(t, !0);
					V(() => {
						Lr(t, "href", W(n).url), q(r, W(n).ref);
					}), Qn("click", t, (e) => e.stopPropagation()), K(e, t);
				}, a = /* @__PURE__ */ it(() => W(n).url && Qr(W(n).url)), o = (e) => {
					var t = sr();
					V(() => q(t, W(n).ref)), K(e, t);
				};
				J(r, (e) => {
					W(a) ? e(i) : e(o, -1);
				}), K(e, t);
			};
			J(b, (e) => {
				W(n).ref && e(S);
			}), O(v);
			var C = B(v, 2), ee = (e) => {
				var t = Ws(), r = z(t, !0);
				V(() => q(r, W(n).summary)), K(e, t);
			};
			J(C, (e) => {
				W(n).summary && e(ee);
			});
			var te = B(C, 2), ne = R(te), w = (e) => {
				var t = Gs(), r = B(R(t), 2);
				Ar(r, () => Os(W(n).context), !0), O(r), O(t), K(e, t);
			};
			J(ne, (e) => {
				W(n).context && e(w);
			});
			var re = B(ne, 2), ie = (e) => {
				var t = Ks(), n = z(t);
				V(() => q(n, `${W(i).label ?? ""}…`)), K(e, t);
			}, ae = (e) => {
				var i = tc(), a = Gt(i), o = (e) => {
					var t = qs(), n = z(t);
					V(() => q(n, `still listed after ${u[W(r)] ?? ""}`)), K(e, t);
				};
				J(a, (e) => {
					u[W(r)] && e(o);
				});
				var c = B(a, 2), l = (e) => {
					var t = Zs();
					wr(t, 21, () => W(n).options, (e) => e.id, (e, t) => {
						var i = Xs();
						let a;
						var o = R(i), s = R(o), c = B(s), l = (e) => {
							var t = Js();
							t.textContent = " (recommended)", K(e, t);
						};
						J(c, (e) => {
							W(t).recommended && e(l);
						}), O(o);
						var u = B(o, 2), d = (e) => {
							var n = Ys(), r = z(n, !0);
							V(() => q(r, W(t).description)), K(e, n);
						};
						J(u, (e) => {
							W(t).description && e(d);
						}), O(i), V(() => {
							a = Nr(i, 1, "opt svelte-t4vask", null, a, { primary: W(t).recommended }), i.disabled = W(h) === W(r), Lr(i, "data-option", W(t).id), q(s, `${W(t).id ?? ""}. ${W(t).title ?? ""}`);
						}), Qn("click", i, () => g(W(n), _(W(n), W(t)))), K(e, i);
					}), O(t), K(e, t);
				};
				J(c, (e) => {
					W(n).options?.length && e(l);
				});
				var d = B(c, 2);
				wr(d, 17, () => Wr(t.api, W(n).kind, W(n).id), (e) => e.label, (e, t) => {
					var i = Qs(), a = z(i, !0);
					V(() => {
						i.disabled = W(h) === W(r), q(a, W(t).label);
					}), Qn("click", i, () => g(W(n), W(t))), K(e, i);
				});
				var f = B(d, 2), p = (e) => {
					var t = $s();
					V(() => t.disabled = W(s) === W(r)), Qn("click", t, () => m(W(n))), K(e, t);
				}, v = (e) => {
					var t = ec();
					V(() => t.disabled = W(s) === W(r)), Qn("click", t, () => m(W(n))), K(e, t);
				};
				J(f, (e) => {
					W(n).kind === "question" ? e(p) : e(v, -1);
				}), K(e, i);
			};
			J(re, (e) => {
				W(i) ? e(ie) : e(ae, -1);
			}), O(te), O(f);
			var oe = B(f), se = z(oe, !0), ce = z(B(oe), !0);
			O(a), V((e) => {
				o = Nr(a, 1, "click svelte-t4vask", null, o, {
					busy: W(s) === W(r) || !!W(i),
					pending: !!W(i)
				}), Lr(a, "data-item", W(n).kind + ":" + W(n).id), q(d, W(n).kind), q(p, W(n).title || W(n).id), q(y, W(n).id), q(se, W(n).why), q(ce, e);
			}, [() => x(W(n).since)]), Qn("click", a, () => m(W(n))), Qn("click", te, (e) => e.stopPropagation()), K(e, a);
		}), O(a), O(i), O(r), K(e, r);
	};
	J(fe, (e) => {
		W(r) ? e(pe) : W(n) ? W(n).length === 0 ? e(he, 2) : e(ge, -1) : e(me, 1);
	});
	var T = B(fe, 4), _e = (e) => {
		var n = mc();
		wr(n, 21, () => W(i), ({ run: e, steps: t, comment: n }) => e.id, (e, n) => {
			let r = () => W(n).run, i = () => W(n).steps, a = () => W(n).comment, o = /* @__PURE__ */ it(() => C(i())), s = /* @__PURE__ */ it(() => i().find((e) => e.status === "in_progress")), c = /* @__PURE__ */ it(() => ne(W(s)));
			var l = pc(), u = R(l), d = B(R(u), 2), f = z(d, !0), p = B(d, 2), m = z(p, !0), h = z(B(p, 2), !0);
			O(u);
			var g = B(u, 2), _ = R(g), v = (e) => {
				var t = ic(), n = Gt(t), r = z(n, !0);
				Fs(B(n, 2), {
					get h() {
						return W(c);
					},
					onopen: w
				}), V((e) => q(r, e), [() => S(W(s), "fields").step || W(s).title]), K(e, t);
			}, y = (e) => {
				K(e, ac());
			};
			J(_, (e) => {
				W(c) ? e(v) : e(y, -1);
			}), O(g);
			var b = B(g, 2), re = R(b);
			wr(re, 17, () => W(o).slice(0, 3), ({ s: e, r: t }) => e.id, (e, t) => {
				let n = () => W(t).s, r = () => W(t).r, i = /* @__PURE__ */ it(() => ne(n()));
				var a = cc(), o = R(a), s = z(o, !0), c = B(o, 2), l = z(c, !0), u = B(c, 2), d = (e) => {
					Fs(e, {
						get h() {
							return W(i);
						},
						onopen: w
					});
				}, f = (e) => {
					var t = oc(), n = z(t, !0);
					V(() => {
						Lr(t, "href", "?session=" + r().by), q(n, r().by);
					}), Qn("click", t, (e) => {
						e.preventDefault(), w(r().by);
					}), K(e, t);
				}, p = /* @__PURE__ */ it(() => te(r().by));
				J(u, (e) => {
					W(i) ? e(d) : W(p) && e(f, 1);
				});
				var m = B(u, 2), h = (e) => {
					var t = sc(), n = z(t, !0);
					V(() => q(n, r().note)), K(e, t);
				};
				J(m, (e) => {
					r().note && e(h);
				});
				var g = z(B(m, 2), !0);
				O(a), V((e, t, n) => {
					Nr(o, 1, `pill ${e ?? ""}`, "svelte-t4vask"), q(s, r().state), q(l, t), q(g, n);
				}, [
					() => ee(r().state),
					() => S(n(), "fields").step || n().title,
					() => x(r().at)
				]), K(e, a);
			}, (e) => {
				var t = cr(), n = Gt(t), r = (e) => {
					var t = cr(), n = Gt(t), r = (e) => {
						var t = lc(), n = R(t), r = z(n, !0), i = z(B(n), !0);
						k(), O(t), V((e) => {
							q(r, W(s).status), q(i, e);
						}, [() => S(W(s), "fields").step || W(s).title]), K(e, t);
					};
					J(n, (e) => {
						W(c) || e(r);
					}), K(e, t);
				}, i = (e) => {
					K(e, uc());
				};
				J(n, (e) => {
					W(s) ? e(r) : e(i, -1);
				}), K(e, t);
			});
			var ie = B(re, 2), ae = (e) => {
				var n = fc(), r = B(R(n)), i = R(r), o = (e) => {
					var n = dc(), r = z(n, !0);
					V(() => {
						Lr(n, "href", "?session=" + a().author), q(r, a().author);
					}), Qn("click", n, (e) => {
						e.preventDefault(), t.api.openSession(a().author);
					}), K(e, n);
				}, s = /* @__PURE__ */ it(() => te(a().author)), c = (e) => {
					var t = sr();
					V(() => q(t, a().author)), K(e, t);
				};
				J(i, (e) => {
					W(s) ? e(o) : e(c, -1);
				}), O(r);
				var l = B(r), u = z(l, !0), d = z(B(l), !0);
				O(n), V((e) => {
					q(u, a().body), q(d, e);
				}, [() => x(a().created_at)]), K(e, n);
			};
			J(ie, (e) => {
				a() && e(ae);
			}), O(b), O(l), V((e) => {
				Lr(l, "data-run", r().id), q(f, r().title), q(m, r().id), q(h, e);
			}, [() => x(r().updated_at)]), Qn("click", u, () => t.api.openTicket(r().id)), K(e, l);
		}), O(n), K(e, n);
	}, ve = (e) => {
		K(e, hc());
	}, ye = (e) => {
		K(e, gc());
	};
	J(T, (e) => {
		W(i).length ? e(_e) : W(o) ? e(ye, -1) : e(ve, 1);
	}), O(le), K(e, re), Ve();
}
$n(["click"]);
// plugins/attention/index.svelte.js
var yc = new URL("./style.css", import.meta.url).href, bc = class extends HTMLElement {
	#e = /* @__PURE__ */ P(Rt({
		api: null,
		context: {},
		config: {}
	}));
	get props() {
		return W(this.#e);
	}
	set props(e) {
		F(this.#e, e, !0);
	}
	app = null;
	set api(e) {
		this.props.api = e;
	}
	get api() {
		return this.props.api;
	}
	set context(e) {
		this.props.context = e;
	}
	get context() {
		return this.props.context;
	}
	set config(e) {
		this.props.config = e;
	}
	get config() {
		return this.props.config;
	}
	connectedCallback() {
		if (!document.querySelector(`link[href="${yc}"]`)) {
			let e = document.createElement("link");
			e.rel = "stylesheet", e.href = yc, document.head.append(e);
		}
		this.style.display = "contents", this.app = hr(vc, {
			target: this,
			props: this.props
		});
	}
	disconnectedCallback() {
		this.app && yr(this.app), this.app = null;
	}
};
customElements.get("desk-attention-view") || customElements.define("desk-attention-view", bc);
