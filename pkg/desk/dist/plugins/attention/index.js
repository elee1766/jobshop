// node_modules/svelte/src/constants.js
var e = {}, t = Symbol("uninitialized"), n = "http://www.w3.org/1999/xhtml", r = Array.isArray, i = Array.prototype.indexOf, a = Array.prototype.includes, o = Array.from, s = Object.defineProperty, c = Object.getOwnPropertyDescriptor, l = Object.getOwnPropertyDescriptors, u = Object.prototype, d = Array.prototype, f = Object.getPrototypeOf, p = Object.isExtensible, m = () => {};
function h(e) {
	for (var t = 0; t < e.length; t++) e[t]();
}
function g() {
	var e, t;
	return {
		promise: new Promise((n, r) => {
			e = n, t = r;
		}),
		resolve: e,
		reject: t
	};
}
var _ = 1024, v = 2048, y = 4096, b = 8192, x = 16384, S = 32768, C = 1 << 25, w = 65536, ee = 1 << 19, te = 1 << 20, T = 1 << 25, ne = 1 << 21, re = 1 << 22, ie = 1 << 23, ae = Symbol("$state"), oe = Symbol("component"), se = Symbol("legacy props"), ce = Symbol(""), le = Symbol("attributes"), ue = Symbol("class"), de = Symbol("style"), fe = Symbol("text"), pe = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), me = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function he() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function E(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function ge() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var D = !1;
function _e(e) {
	D = e;
}
var O;
function ve(t) {
	if (t === null) throw E(), e;
	return O = t;
}
function ye() {
	return ve(/* @__PURE__ */ Ht(O));
}
function k(t) {
	if (D) {
		if (/* @__PURE__ */ Ht(O) !== null) throw E(), e;
		O = t;
	}
}
function be(e = 1) {
	if (D) {
		for (var t = e, n = O; t--;) n = /* @__PURE__ */ Ht(n);
		O = n;
	}
}
function xe(e = !0) {
	for (var t = 0, n = O;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ Ht(n);
		e && n.remove(), n = i;
	}
}
function A(t) {
	if (!t || t.nodeType !== 8) throw E(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function Se(e) {
	return e === this.v;
}
function j(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function Ce(e) {
	return !j(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function we() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Te(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function Ee(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function De() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function Oe(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function ke() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Ae(e) {
	throw Error("https://svelte.dev/e/props_invalid_value");
}
function je() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Me() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Ne() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Pe() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var Fe = null;
function Ie(e) {
	Fe = e;
}
function Le(e, t = !1, n) {
	Fe = {
		p: Fe,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: U,
		l: null
	};
}
function Re(e) {
	var t = Fe, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) nn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, Fe = t.p, ze(e);
}
function ze(e = {}) {
	return s(e, oe, { value: !0 }), e;
}
function Be() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var Ve = [];
function He() {
	var e = Ve;
	Ve = [], h(e);
}
function Ue(e) {
	if (Ve.length === 0 && !ft) {
		var t = Ve;
		queueMicrotask(() => {
			t === Ve && He();
		});
	}
	Ve.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var We = ~(v | y | _);
function M(e, t) {
	e.f = e.f & We | t;
}
function Ge(e) {
	e.f & 512 || e.deps === null ? M(e, _) : M(e, y);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Ke(e, t) {
	e.f & 2048 ? t.set(e, v) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, y), M(e, _);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function qe(e) {
	var t = H, n = U;
	wn(null), Tn(null);
	try {
		return e();
	} finally {
		wn(t), Tn(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function Je(e, t, n, r) {
	let i = Be() ? Qe : nt;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = U, c = Ye(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				Yt(e, s);
			}
			Xe();
		}
	}
	var d = Ze();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ et(e))).then(u).catch((e) => Yt(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), Xe());
	}) : f();
}
function Ye() {
	var e = U, t = H, n = Fe, r = N;
	return function(i = !0) {
		Tn(e), wn(t), Ie(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function Xe(e = !0) {
	Tn(null), wn(null), Ie(null), e && N?.deactivate();
}
function Ze() {
	var e = U, t = e.b, n = N, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Qe(e) {
	var n = 2 | v;
	return U !== null && (U.f |= ee), {
		ctx: Fe,
		deps: null,
		effects: null,
		equals: Se,
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
var $e = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function et(e, n, r) {
	let i = U;
	i === null && we();
	var a = void 0, o = Dt(t), s = !H, c = /* @__PURE__ */ new Set();
	return an(() => {
		var t = U, n = g();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== pe && n.reject(e);
			}).finally(Xe);
		} catch (e) {
			n.reject(e), Xe();
		}
		var r = N;
		if (s) {
			if (t.f & 32768) var l = Ze();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject($e);
			else for (let e of c.values()) e.reject($e);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== $e && (r.activate(), t ? (o.f |= ie, jt(o, t)) : (o.f & 8388608 && (o.f ^= ie), jt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), en(() => {
		for (let e of c) e.reject($e);
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
function tt(e) {
	let t = /* @__PURE__ */ Qe(e);
	return Dn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function nt(e) {
	let t = /* @__PURE__ */ Qe(e);
	return t.equals = Ce, t;
}
function rt(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) fn(t[n]);
	}
}
function it(e) {
	var n, r = U, i = e.parent;
	if (!xn && i !== null && e.v !== t && i.f & 24576) return he(), e.v;
	Tn(i);
	try {
		rt(e), n = zn(e);
	} finally {
		Tn(r);
	}
	return n;
}
function at(e) {
	var t = it(e);
	!e.equals(t) && (e.wv = In(), (!N?.is_fork || e.deps === null) && (N === null ? e.v = t : (N.capture(e, t, !0), lt?.capture(e, t, !0)), e.deps === null)) ? M(e, _) : xn || (ut === null ? Ge(e) : ($t() || N?.is_fork) && ut.set(e, t));
}
function ot(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && qe(() => {
		t.ac.abort(pe), t.ac = null;
	}), t.fn !== null && (t.teardown = m), Hn(t, 0), un(t));
}
function st(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Un(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var ct = null, N = null, lt = null, ut = null, dt = null, ft = !1, P = !1, pt = null, mt = null, ht = 0, gt = 1, _t = class e {
	id = gt++;
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
		ct === null ? ct = this : (ct.next = this, this.prev = ct), ct = this;
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
			for (var r of n.d) M(r, v), t(r);
			for (r of n.m) M(r, y), t(r);
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
					t.f ^= _;
				}
			}
			n || e.push(t);
		}
		return this.#a = [], e;
	}
	#m() {
		this.started = !0;
		for (let [e, t] of this.#s) M(e, t), this.schedule(e);
		this.apply();
		for (var t = pt = [], n = [], r = mt = []; this.#a.length > 0;) {
			ht++ > 1e3 && (this.#b(), vt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Ct(e), this.#d() || this.discard(), t;
			}
		}
		if (N = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (pt = null, mt = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) St(e, t);
			r.length > 0 && N.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), lt = this, bt(n), bt(t), lt = null, this.#i?.resolve();
			var o = N;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Tt.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= _;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= _ : i & 4 ? t.push(r) : Ln(r) && (i & 16 && this.add_dirty_reaction(r, y), Un(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), M(i, v), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), N = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) Ke(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), ut?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		N = this;
	}
	deactivate() {
		N = null, ut = null;
	}
	flush() {
		try {
			P = !0, N = this, this.#m();
		} finally {
			ht = 0, dt = null, pt = null, mt = null, P = !1, N = null, ut = null, Tt.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject($e);
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
		this.#u || (this.#u = !0, Ue(() => {
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
		return (this.#i ??= g()).promise;
	}
	static ensure() {
		if (N === null) {
			let t = N = new e();
			!P && Ue(() => {
				t.started || t.flush();
			});
		}
		return N;
	}
	apply() {
		ut = null;
	}
	schedule(e) {
		dt = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? ct = e : t.prev = e, this.linked = !1;
		}
	}
};
function vt() {
	try {
		ke();
	} catch (e) {
		Yt(e, dt);
	}
}
var yt = null;
function bt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Ln(r) && (yt = /* @__PURE__ */ new Set(), Un(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && mn(r), yt?.size > 0)) {
				Tt.clear();
				for (let e of yt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) yt.has(n) && (yt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Un(n);
					}
				}
				yt.clear();
			}
		}
		yt = null;
	}
}
function xt(e) {
	N.schedule(e);
}
function St(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), M(e, _);
		for (var n = e.first; n !== null;) St(n, t), n = n.next;
	}
}
function Ct(e) {
	M(e, _);
	for (var t = e.first; t !== null;) Ct(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var wt = /* @__PURE__ */ new Set(), Tt = /* @__PURE__ */ new Map(), Et = !1;
function Dt(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: Se,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function F(e, t) {
	let n = Dt(e, t);
	return Dn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Ot(e, t = !1, n = !0) {
	let r = Dt(e);
	return t || (r.equals = Ce), r;
}
function I(e, t, n = !1) {
	return H !== null && (!Cn || H.f & 131072) && Be() && H.f & 4325394 && (En === null || !En.has(e)) && Ne(), jt(e, n ? Ft(t) : t, mt);
}
var kt = null, At = 0;
function jt(e, t, n = null) {
	if (!e.equals(t)) {
		xn ? Tt.set(e, t) : Tt.has(e) || Tt.set(e, e.v);
		var r = _t.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && it(t), ut === null && Ge(t);
		}
		e.wv = In(), kt = null, At = 0, Pt(e, v, n), kt = null, Be() && U !== null && U.f & 1024 && !(U.f & 96) && (An === null ? jn([e]) : An.push(e)), !r.is_fork && wt.size > 0 && !Et && Mt();
	}
	return t;
}
function Mt() {
	Et = !1;
	for (let e of wt) {
		e.f & 1024 && M(e, y);
		let t;
		try {
			t = Ln(e);
		} catch {
			t = !0;
		}
		t && Un(e);
	}
	wt.clear();
}
function Nt(e) {
	I(e, e.v + 1);
}
function Pt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Be(), a = r.length;
		if (At += a, At > 1e5 && kt === null && (kt = /* @__PURE__ */ new Set()), kt !== null) {
			if (kt.has(e)) return;
			kt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== U) {
				var l = (c & v) === 0;
				if (l && M(s, t), c & 131072) wt.add(s);
				else if (c & 2) {
					var u = s;
					ut?.delete(u), Pt(u, y, n);
				} else if (l) {
					var d = s;
					c & 16 && yt !== null && yt.add(d), n === null ? xt(d) : n.push(d);
				}
			}
		}
	}
}
function Ft(e) {
	if (typeof e != "object" || !e || ae in e || oe in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ F(0), s = null, l = Pn, p = (e) => {
		if (Pn === l) return e();
		var t = H, n = Pn;
		wn(null), Fn(l);
		var r = e();
		return wn(t), Fn(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ F(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || je();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ F(n.value, s);
				return i.set(t, e), e;
			}) : I(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ F(t, s));
					i.set(n, e), Nt(o);
				}
			} else I(r, t), Nt(o);
			return !0;
		},
		get(n, r, a) {
			if (r === ae) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ F(Ft(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
				var u = W(o);
				return u === t ? void 0 : u;
			}
			return Reflect.get(n, r, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var r = Reflect.getOwnPropertyDescriptor(e, n), a = i.get(n);
			if (a !== void 0) {
				var o = W(a);
				if (o === t) return;
				if (r && "value" in r) r.value = o;
				else return {
					enumerable: !0,
					configurable: !0,
					value: o,
					writable: !0
				};
			}
			return r;
		},
		has(e, n) {
			if (n === ae) return !0;
			var r = i.get(n), a = r !== void 0 && r.v !== t || Reflect.has(e, n);
			return (r !== void 0 || (U !== null || H !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ F(a ? Ft(e[n]) : t, s)), i.set(n, r)), W(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ F(t, s)), i.set(f + "", m)) : I(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ F(void 0, s)), I(u, Ft(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var h = p(() => Ft(r));
				I(u, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && I(_, v + 1);
				}
				Nt(o);
			}
			return !0;
		},
		ownKeys(e) {
			W(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = i.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [r, a] of i) a.v !== t && !(r in e) && n.push(r);
			return n;
		},
		setPrototypeOf() {
			Me();
		}
	});
}
var It, Lt, Rt, zt;
function Bt() {
	if (It === void 0) {
		It = window, Lt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Rt = c(t, "firstChild").get, zt = c(t, "nextSibling").get, p(e) && (e[ue] = void 0, e[le] = null, e[de] = void 0, e.__e = void 0), p(n) && (n[fe] = void 0);
	}
}
function L(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Vt(e) {
	return Rt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Ht(e) {
	return zt.call(e);
}
function R(e, t) {
	if (!D) return /* @__PURE__ */ Vt(e);
	var n = /* @__PURE__ */ Vt(O);
	if (n === null) n = O.appendChild(L());
	else if (t && n.nodeType !== 3) {
		var r = L();
		return n?.before(r), ve(r), r;
	}
	return t && qt(n), ve(n), n;
}
function Ut(e, t = !1) {
	if (!D) {
		var n = /* @__PURE__ */ Vt(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Ht(n) : n;
	}
	if (t) {
		if (O?.nodeType !== 3) {
			var r = L();
			return O?.before(r), ve(r), r;
		}
		qt(O);
	}
	return O;
}
function z(e, t = !1) {
	if (!D) return /* @__PURE__ */ Vt(e);
	var n = R(e, t);
	return k(e), n;
}
function B(e, t = 1, n = !1) {
	let r = D ? O : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Ht(r);
	if (!D) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = L();
			return r === null ? i?.after(a) : r.before(a), ve(a), a;
		}
		qt(r);
	}
	return ve(r), r;
}
function Wt(e) {
	e.textContent = "";
}
function Gt() {
	return !1;
}
function Kt(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function qt(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function Jt(e) {
	var t = U;
	if (t === null) return H.f |= ie, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	Yt(e, t);
}
function Yt(t, n) {
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
function Xt(e) {
	U === null && (H === null && Oe(e), De()), xn && Ee(e);
}
function Zt(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function Qt(e, t) {
	var n = U;
	n !== null && n.f & 8192 && (e |= b);
	var r = {
		ctx: Fe,
		deps: null,
		nodes: null,
		f: e | v | 512,
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
	N?.register_created_effect(r);
	var i = r;
	if (e & 4) pt === null ? _t.ensure().schedule(r) : pt.push(r);
	else if (t !== null) {
		try {
			Un(r);
		} catch (e) {
			throw fn(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= w));
	}
	if (i !== null && (i.parent = n, n !== null && Zt(i, n), H !== null && H.f & 2 && !(e & 64))) {
		var a = H;
		(a.effects ??= []).push(i);
	}
	return r;
}
function $t() {
	return H !== null && !Cn;
}
function en(e) {
	let t = Qt(8, null);
	return M(t, _), t.teardown = e, t;
}
function tn(e) {
	Xt("$effect");
	var t = U.f;
	if (!H && t & 32 && Fe !== null && !Fe.i) {
		var n = Fe;
		(n.e ??= []).push(e);
	} else return nn(e);
}
function nn(e) {
	return Qt(4 | te, e);
}
function rn(e) {
	_t.ensure();
	let t = Qt(64 | ee, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? hn(t, () => {
			fn(t), n(void 0);
		}) : (fn(t), n(void 0));
	});
}
function an(e) {
	return Qt(re | ee, e);
}
function on(e, t = 0) {
	return Qt(8 | t, e);
}
function V(e, t = [], n = [], r = []) {
	Je(r, t, n, (t) => {
		Qt(8, () => {
			e(...t.map(W));
		});
	});
}
function sn(e, t = 0) {
	return Qt(16 | t, e);
}
function cn(e) {
	return Qt(32 | ee, e);
}
function ln(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = xn, r = H;
		Sn(!0), wn(null);
		try {
			t.call(null);
		} catch (t) {
			Yt(t, e.parent);
		} finally {
			Sn(n), wn(r);
		}
	}
}
function un(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && qe(() => {
			e.abort(pe);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : fn(n, t), n = r;
	}
}
function dn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || fn(t), t = n;
	}
}
function fn(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (pn(e.nodes.start, e.nodes.end), n = !0), e.f |= C, un(e, t && !n), Hn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	ln(e), e.f ^= C, e.f |= x;
	var i = e.parent;
	i !== null && i.first !== null && mn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function pn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Ht(e);
		e.remove(), e = n;
	}
}
function mn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function hn(e, t, n = !0) {
	var r = [];
	e.f |= 256, gn(e, r, !0);
	var i = () => {
		n && fn(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function gn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= b;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				gn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function _n(e) {
	e.f &= -257, vn(e, !0);
}
function vn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= b, e.f & 1024 || (M(e, v), _t.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			vn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function yn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ Ht(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var bn = null, xn = !1;
function Sn(e) {
	xn = e;
}
var H = null, Cn = !1;
function wn(e) {
	H = e;
}
var U = null;
function Tn(e) {
	U = e;
}
var En = null;
function Dn(e) {
	H !== null && (H.f & 2097152 || H.f & 2) && (En ??= /* @__PURE__ */ new Set()).add(e);
}
var On = null, kn = 0, An = null;
function jn(e) {
	An = e;
}
var Mn = 1, Nn = 0, Pn = Nn;
function Fn(e) {
	Pn = e;
}
function In() {
	return ++Mn;
}
function Ln(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Ln(a) && at(a), a.wv > e.wv) return !0;
		}
		t & 512 && ut === null && M(e, _);
	}
	return !1;
}
function Rn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(En !== null && En.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Rn(a, t, !1) : t === a && (n ? M(a, v) : a.f & 1024 && M(a, y), xt(a));
	}
}
function zn(e) {
	var t = On, n = kn, r = An, i = H, a = En, o = Fe, s = Cn, c = Pn, l = e.f;
	On = null, kn = 0, An = null, H = l & 96 ? null : e, En = null, Ie(e.ctx), Cn = !1, Pn = ++Nn, e.ac !== null && (qe(() => {
		e.ac.abort(pe);
	}), e.ac = null);
	try {
		e.f |= ne;
		var u = e.fn, d = u();
		e.f |= S;
		var f = Bn(e);
		if (Be() && An !== null && !Cn && f !== null && !(e.f & 6146)) for (var p = 0; p < An.length; p++) Rn(An[p], e);
		if (i !== null && i !== e) {
			if (Nn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Nn;
			if (t !== null) for (let e of t) e.rv = Nn;
			An !== null && (r === null ? r = An : r.push(...An));
		}
		return e.f & 8388608 && (e.f ^= ie), d;
	} catch (t) {
		return Bn(e), Jt(t);
	} finally {
		e.f ^= ne, On = t, kn = n, An = r, H = i, En = a, Ie(o), Cn = s, Pn = c;
	}
}
function Bn(e) {
	var t = e.deps, n = N?.is_fork;
	if (On !== null) {
		var r;
		if (n || Hn(e, kn), t !== null && kn > 0) for (t.length = kn + On.length, r = 0; r < On.length; r++) t[kn + r] = On[r];
		else e.deps = t = On;
		if ($t() && e.f & 512) for (r = kn; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && kn < t.length && (Hn(e, kn), t.length = kn);
	return t;
}
function Vn(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var o = i.call(r, e);
		if (o !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[o] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (On === null || !a.call(On, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Ge(c), c.ac !== null && qe(() => {
			c.ac.abort(pe), c.ac = null, M(c, v);
		}), ot(c), Hn(c, 0);
	}
}
function Hn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Vn(e, n[r]);
}
function Un(e) {
	var t = e.f;
	if (!(t & 16384)) {
		M(e, _);
		var n = U;
		U = e;
		try {
			t & 16777232 ? dn(e) : un(e), ln(e);
			var r = zn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Mn;
		} finally {
			U = n;
		}
	}
}
function W(e) {
	var t = !!(e.f & 2);
	if (bn?.add(e), H !== null && !Cn && !(U !== null && U.f & 16384) && (En === null || !En.has(e))) {
		var n = H.deps;
		if (H.f & 2097152) e.rv < Nn && (e.rv = Nn, On === null && n !== null && n[kn] === e ? kn++ : On === null ? On = [e] : On.push(e));
		else {
			H.deps ??= [], a.call(H.deps, e) || H.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [H] : a.call(r, H) || r.push(H);
		}
	}
	if (xn && Tt.has(e)) return Tt.get(e);
	if (t) {
		var i = e;
		if (xn) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || Gn(i)) && (o = it(i)), Tt.set(i, o), o;
		}
		var s = !(i.f & 512) && !Cn && H !== null && !!(H.f & 512), c = (i.f & S) === 0;
		Ln(i) && (s && (i.f |= 512), at(i)), s && !c && (st(i), Wn(i));
	}
	if (ut?.has(e)) return ut.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function Wn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (st(n), Wn(n));
	}
}
function Gn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Tt.has(t) || t.f & 2 && Gn(t)) return !0;
	return !1;
}
function Kn(e) {
	var t = Cn;
	try {
		return Cn = !0, e();
	} finally {
		Cn = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var qn = Symbol("events"), Jn = /* @__PURE__ */ new Set(), Yn = /* @__PURE__ */ new Set();
function Xn(e, t, n) {
	(t[qn] ??= {})[e] = n;
}
function Zn(e) {
	for (var t = 0; t < e.length; t++) Jn.add(e[t]);
	for (var n of Yn) n(e);
}
var Qn = null, $n = !1;
function er(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	Qn = e, $n || ($n = !0, setTimeout(() => {
		$n = !1, Qn = null;
	}));
	var o = 0, c = Qn === e && e[qn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[qn] = t;
			return;
		}
		var u = i.indexOf(t);
		if (u === -1) return;
		l <= u && (o = l);
	}
	if (a = i[o] || e.target, a !== t) {
		s(e, "currentTarget", {
			configurable: !0,
			get() {
				return a || n;
			}
		});
		var d = H, f = U;
		wn(null), Tn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[qn]?.[r];
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
			e[qn] = t, delete e.currentTarget, wn(d), Tn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var tr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function nr(e) {
	return tr?.createHTML(e) ?? e;
}
function rr(e) {
	var t = Kt("template");
	return t.innerHTML = nr(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function ir(e, t) {
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
		if (D) return ir(O, null), O;
		i === void 0 && (i = rr(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Vt(i)));
		var t = r || Lt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Vt(t), s = t.lastChild;
			ir(o, s);
		} else ir(t, t);
		return t;
	};
}
function ar(e = "") {
	if (!D) {
		var t = L(e + "");
		return ir(t, t), t;
	}
	var n = O;
	return n.nodeType === 3 ? qt(n) : (n.before(n = L()), ve(n)), ir(n, n), n;
}
function or() {
	if (D) return ir(O, null), O;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = L();
	return e.append(t, n), ir(t, n), e;
}
function K(e, t) {
	if (D) {
		var n = U;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = O), ye();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var sr = ["touchstart", "touchmove"];
function cr(e) {
	return sr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function lr(e) {
	let t = 0, n = Dt(0), r;
	return () => {
		$t() && (W(n), on(() => (t === 0 && (r = Kn(() => e(() => Nt(n)))), t += 1, () => {
			Ue(() => {
				--t, t === 0 && (r?.(), r = void 0, Nt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var ur = w | ee;
function dr(e, t, n, r) {
	new fr(e, t, n, r);
}
var fr = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = D ? O : null;
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
	#m = lr(() => (this.#p = Dt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = U;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = U.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = sn(() => {
			if (D) {
				let e = A(this.#t);
				ye();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, ur), D && (this.#e = O);
	}
	#h() {
		try {
			this.#a = cn(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		Ue(r), t && (this.#s = cn(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? ge() : (t = !0, n && Pe(), this.#s !== null && hn(this.#s, () => {
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
					Yt(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (x | C)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = cn(() => e(this.#e)), Ue(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = L(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return cn(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							Yt(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(N)) : this.#u === 0 && (this.#e.before(e), this.#c = null, hn(this.#o, () => {
					this.#o = null;
				}), this.#x(N));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = cn(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				yn(this.#a, e);
				let t = this.#n.pending;
				this.#o = cn(() => t(this.#e));
			} else this.#x(N);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		Ke(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = U, n = H, r = Fe;
		Tn(this.#i), wn(this.#i), Ie(this.#i.ctx);
		try {
			return _t.ensure(), e();
		} finally {
			Tn(t), wn(n), Ie(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && hn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Ue(() => {
			this.#d = !1, this.#p && jt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), W(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		N?.is_fork ? (this.#a && N.skip_effect(this.#a), this.#o && N.skip_effect(this.#o), this.#s && N.skip_effect(this.#s), N.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (fn(this.#a), null), this.#o &&= (fn(this.#o), null), this.#s &&= (fn(this.#s), null), D && (ve(this.#t), be(), ve(xe()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return cn(() => {
						var r = U;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return Yt(e, this.#i.parent), null;
				}
			}));
		};
		Ue(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					Yt(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => Yt(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function q(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[fe] ??= e.nodeValue) && (e[fe] = n, e.nodeValue = `${n}`);
}
function pr(e, t) {
	return hr(e, t);
}
var mr = /* @__PURE__ */ new Map();
function hr(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	Bt();
	var u = void 0, d = rn(() => {
		var c = r ?? n.appendChild(L());
		dr(c, { pending: () => {} }, (n) => {
			Le({});
			var r = Fe;
			if (s && (r.c = s), a && (i.$$events = a), D && ir(n, null), u = t(n, i) || ze(), D && (U.nodes.end = O, O === null || O.nodeType !== 8 || O.data !== "]")) throw E(), e;
			Re();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = cr(r);
					for (let e of [n, document]) {
						var a = mr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), mr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, er, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(Jn)), Yn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = mr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, er), t.delete(e), t.size === 0 && mr.delete(r)) : t.set(e, i);
			}
			Yn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return gr.set(u, d), u;
}
var gr = /* @__PURE__ */ new WeakMap();
function _r(e, t) {
	let n = gr.get(e);
	return n ? (gr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var vr = class {
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
			if (n) _n(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (_n(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (fn(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						yn(r, t), t.append(L()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else fn(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), hn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (fn(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = N, r = Gt();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = L();
				i.append(a), this.#n.set(e, {
					effect: cn(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, cn(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else D && (this.anchor = O), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function yr(e, t, n = !1) {
	var r;
	D && (r = O, ye());
	var i = new vr(e), a = n ? w : 0;
	function o(e, t) {
		if (D) {
			var n = A(r);
			if (e !== parseInt(n.substring(1))) {
				var a = xe();
				ve(a), i.anchor = a, _e(!1), i.ensure(e, t), _e(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	sn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function br(e, t, n) {
	for (var r = [], i = t.length, a, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		hn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					xr(e, o(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --s;
		}, !1);
	}
	if (s === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			Wt(d), d.append(u), e.items.clear();
		}
		xr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function xr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= T, yn(a, document.createDocumentFragment())) : fn(t[i], n);
	}
}
var Sr;
function Cr(e, t, n, i, a, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = D ? ve(/* @__PURE__ */ Vt(u)) : u.appendChild(L());
	}
	D && ye();
	var d = null, f = /* @__PURE__ */ nt(() => {
		var e = n();
		return r(e) ? e : e == null ? [] : o(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = W(f);
			_.fallback = d, Tr(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= T, Dr(d, null, c)) : _n(d) : hn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: sn(() => {
			var e = W(f), r = e.length;
			let o = !1;
			D && A(c) === "[!" != (r === 0) && (c = xe(), ve(c), _e(!1), o = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = N, v = Gt(), y = 0; y < r; y += 1) {
				D && O.nodeType === 8 && O.data === "]" && (c = O, o = !0, _e(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && jt(S.v, b), S.i && jt(S.i, y), v && _.unskip_effect(S.e)) : (S = Er(l, m ? c : Sr ??= L(), b, x, y, a, t, n), m || (S.e.f |= T), l.set(x, S)), u.add(x);
			}
			if (r === 0 && s && !d && (m ? d = cn(() => s(c)) : (d = cn(() => s(Sr ??= L())), d.f |= T)), r > u.size && Te("", "", ""), D && r > 0 && ve(xe()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			o && _e(!0), W(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, D && (c = O);
}
function wr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function Tr(e, t, n, r, i) {
	var a = !!(r & 8), s = t.length, c = e.items, l = wr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (_n(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= T, _ === l) Dr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), Or(e, d, _), Or(e, _, y), Dr(_, y, n), d = _, p = [], m = [], l = wr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) Dr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					Or(e, S.prev, C.next), Or(e, d, S), Or(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), Dr(_, l, n), Or(e, _.prev, _.next), Or(e, _, d === null ? e.effect.first : d.next), Or(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = wr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = wr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (xr(e, o(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var w = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || w.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && w.push(l), l = wr(l.next);
		var ee = w.length;
		if (ee > 0) {
			var te = r & 4 && s === 0 ? n : null;
			if (a) {
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.measure();
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.fix();
			}
			br(e, w, te);
		}
	}
	a && Ue(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Er(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Dt(n) : /* @__PURE__ */ Ot(n, !1, !1) : null, l = o & 2 ? Dt(i) : null;
	return {
		v: c,
		i: l,
		e: cn(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Dr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Ht(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function Or(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/shared/attributes.js
var kr = [..." 	\n\r\f\xA0\v﻿"];
function Ar(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || kr.includes(r[o - 1])) && (s === r.length || kr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function jr(e, t, n, r, i, a) {
	var o = e[ue];
	if (D || o !== n || o === void 0) {
		var s = Ar(n, r, a);
		D && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[ue] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Mr = Symbol("is custom element"), Nr = Symbol("is html"), Pr = me ? "link" : "LINK";
function Fr(e, t, n, r) {
	var i = Ir(e);
	if (!(D && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Pr))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[ce] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Rr(e).has(t) ? e[t] = n : (!D || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Ir(e) {
	return e[le] ??= {
		[Mr]: e.nodeName.includes("-"),
		[Nr]: e.namespaceURI === n
	};
}
var Lr = /* @__PURE__ */ new Map();
function Rr(e) {
	var t = e.getAttribute("is") || e.nodeName, n = Lr.get(t);
	if (n) return n;
	Lr.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = l(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = f(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var zr = !1;
function Br(e) {
	var t = zr;
	try {
		return zr = !1, [e(), zr];
	} finally {
		zr = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
function Vr(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, l = !0, u = void 0, d = () => o && i ? (u ??= /* @__PURE__ */ Qe(r), W(u)) : (l && (l = !1, s = o ? Kn(r) : r), s);
	let f;
	if (a) {
		var p = ae in e || se in e;
		f = c(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = Br(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Ae(t), f(m)));
	var g = i ? () => {
		var n = e[t];
		return n === void 0 ? d() : (l = !0, n);
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
	var v = !1, y = (n & 1 ? Qe : nt)(() => (v = !1, g()));
	a && W(y);
	var b = U;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? W(y) : i && a ? Ft(e) : e;
			return I(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return xn && v || b.f & 16384 ? y.v : W(y);
	});
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// kit/attention.js
function Hr(e, t, n) {
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
async function Ur(e, t) {
	if (!await e.confirm(t.ask)) return !1;
	try {
		return await t.run(), e.toast(t.label + ": done"), !0;
	} catch (n) {
		return e.toast(t.label + " failed: " + String(n.message || n)), !1;
	}
}
// node_modules/marked/lib/marked.esm.js
function Wr() {
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
var Gr = Wr();
function Kr(e) {
	Gr = e;
}
var qr = { exec: () => null };
function Jr(e) {
	let t = [];
	return (n) => {
		let r = Math.max(0, Math.min(3, n - 1)), i = t[r];
		return i || (i = e(r), t[r] = i), i;
	};
}
function J(e, t = "") {
	let n = typeof e == "string" ? e : e.source, r = {
		replace: (e, t) => {
			let i = typeof t == "string" ? t : t.source;
			return i = i.replace(Y.caret, "$1"), n = n.replace(e, i), r;
		},
		getRegex: () => new RegExp(n, t)
	};
	return r;
}
var Yr = ((e = "") => {
	try {
		return !!RegExp("(?<=1)(?<!1)" + e);
	} catch {
		return !1;
	}
})(), Y = {
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
	nextBulletRegex: Jr((e) => RegExp(`^ {0,${e}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`)),
	hrRegex: Jr((e) => RegExp(`^ {0,${e}}((?:-[ 	]*){3,}|(?:_[ 	]*){3,}|(?:\\*[ 	]*){3,})(?:\\n+|$)`)),
	fencesBeginRegex: Jr((e) => RegExp(`^ {0,${e}}(?:\`\`\`|~~~)`)),
	headingBeginRegex: Jr((e) => RegExp(`^ {0,${e}}#{1,6}(?:\\s|$)`)),
	pedanticHeadingBeginRegex: Jr((e) => RegExp(`^ {0,${e}}#`)),
	htmlBeginRegex: Jr((e) => RegExp(`^ {0,${e}}(?:</?(?:${ui})(?: +|$|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)`, "i")),
	blockquoteBeginRegex: Jr((e) => RegExp(`^ {0,${e}}>`))
}, Xr = /^(?:[ \t]*(?:\n|$))+/, Zr = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, Qr = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}(?:(?=`)\1`*|(?=~)\1~*) *(?=\n|$)|$)/, $r = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, ei = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, ti = / {0,3}(?:[*+-]|\d{1,9}[.)])/, ni = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |fences|blockquote|heading|hr|html|table))+?)\n {0,3}(=+|-+)[ \t]*(?:\n+|$)/, ri = J(ni).replace(/bull/g, ti).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), ii = J(ni).replace(/bull/g, ti).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), ai = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table|[ \t]+\n)[^\n]+)*)/, oi = /^[^\n]+/, si = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\\n]|\n(?![ \t]*\n))+/, ci = J(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", si).replace("title", /(?:"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\\n]|\n(?!\n))*'|\((?:\\[\s\S]|[^()\\])*\))/).getRegex(), li = J(/^(bull)([ \t][^\n]*?)?(?:\n|$)/).replace(/bull/g, ti).getRegex(), ui = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", di = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, fi = J("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n*|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>[^\\n]*\\n*|$)|<![A-Z][\\s\\S]*?(?:>[^\\n]*\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>[^\\n]*\\n*|$)|</?(tag)(?:[ \\t]+|\\n|/?>|$)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!(?:script|pre|style|textarea)(?![a-z0-9-]))([a-z][a-z0-9-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</[a-z][a-z0-9-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", di).replace("tag", ui).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), pi = (e) => J(ai).replace("hr", $r).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", e).replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", ui).getRegex(), mi = pi(/ {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]/), hi = pi(/ {0,3}(?:[*+-]|\d{1,9}[.)])(?:[ \t]|\n|$)/), gi = {
	blockquote: J(/^( {0,3}> ?(?:(?![ \t]*(?:\n|$))(paragraph)|[^\n]*)(?:\n|$))+/).replace("paragraph", hi).getRegex(),
	code: Zr,
	def: ci,
	fences: Qr,
	heading: ei,
	hr: $r,
	html: fi,
	lheading: ri,
	list: li,
	newline: Xr,
	paragraph: mi,
	table: qr,
	text: oi
}, _i = J("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", $r).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", ui).getRegex(), vi = {
	...gi,
	lheading: ii,
	table: _i,
	paragraph: J(ai).replace("hr", $r).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", _i).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]+[^ \\t\\n]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", ui).getRegex()
}, yi = {
	...gi,
	html: J("^ *(?:comment(?: *\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1>(?: *\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?>(?: *\\n{2,}|\\s*$))").replace("comment", di).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: qr,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: J(ai).replace("hr", $r).replace("heading", " *#{1,6} *[^\n]").replace("lheading", ri).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, bi = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, xi = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, Si = /^( {2,}|\\)\n(?!\s*$)[ \t]*/, Ci = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, wi = /[\p{P}\p{S}]/u, Ti = /[\s\p{P}\p{S}]/u, Ei = /[^\s\p{P}\p{S}]/u, Di = J(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, Ti).getRegex(), Oi = /[\p{Pi}\p{Ps}"']/u, ki = /(?!~)[\p{P}\p{S}]/u, Ai = /(?!~)[\s\p{P}\p{S}]/u, ji = /(?:[^\s\p{P}\p{S}]|~)/u, Mi = J(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", Yr ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), Ni = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, Pi = J(Ni, "u").replace(/punct/g, wi).getRegex(), Fi = J(Ni, "u").replace(/punct/g, ki).getRegex(), Ii = J(/^(?:\*+(?:((?!\*)(?!openQuote)punct)|([^\s*]))?)|^_+(?:((?!_)(?!openQuote)punct)|([^\s_]))?/, "u").replace(/openQuote/g, Oi).replace(/punct/g, wi).getRegex(), Li = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", Ri = J(Li, "gu").replace(/notPunctSpace/g, Ei).replace(/punctSpace/g, Ti).replace(/punct/g, wi).getRegex(), zi = J(Li, "gu").replace(/notPunctSpace/g, ji).replace(/punctSpace/g, Ai).replace(/punct/g, ki).getRegex(), Bi = J("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)[\\s](\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|(?:(?!\\*)punct|notPunctSpace)(\\*+)(?!\\*)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Ei).replace(/punctSpace/g, Ti).replace(/punct/g, wi).getRegex(), Vi = J("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, Ei).replace(/punctSpace/g, Ti).replace(/punct/g, wi).getRegex(), Hi = J("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)[\\s](_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)|(?:(?!_)punct|notPunctSpace)(_+)(?!_)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Ei).replace(/punctSpace/g, Ti).replace(/punct/g, wi).getRegex(), Ui = J(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, wi).getRegex(), Wi = J("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Ei).replace(/punctSpace/g, Ti).replace(/punct/g, wi).getRegex(), Gi = J(/\\(punct)/, "gu").replace(/punct/g, wi).getRegex(), Ki = J(/(`+)(?:[^`]|[^`][\s\S]*?[^`])\1(?!`)|`+|\\(punct)/, "gu").replace(/punct/g, wi).getRegex(), qi = J(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), Ji = J(di).replace("(?:-->|$)", "-->").getRegex(), Yi = J("^comment|^</[a-zA-Z][a-zA-Z0-9-]*\\s*>|^<[a-zA-Z][a-zA-Z0-9-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", Ji).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), Xi = /\[(?:\\[\s\S]|[^\[\]\\])*\]/, Zi = J(/(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/).replace("brackets", Xi).getRegex(), Qi = J(/^!?\[(label)\]\([ \t\n]*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?[ \t\n]*\)/).replace("label", Zi).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]+|(?=\))/).replace("title", /"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|\((?:\\[\s\S]|[^)\\])*\)/).getRegex(), $i = J(/^!?\[(label)\]\[(ref)\]/).replace("label", Zi).replace("ref", si).getRegex(), ea = J(/^!?\[(ref)\](?:\[\])?/).replace("ref", si).getRegex(), ta = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\]){1,999}/, na = J(/(?:[^\[\]\\`]*(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\]))){0,999}?[^\[\]\\`]*?/).replace("brackets", Xi).getRegex(), ra = J("reflink|nolink(?!\\()", "g").replace("reflink", J(/^!?\[(label)\]\[(ref)\]/).replace("label", na).replace("ref", ta).getRegex()).replace("nolink", J(/^!?\[(ref)\](?:\[\])?/).replace("ref", ta).getRegex()).getRegex(), ia = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, aa = J(/(?:mailto:email|xmpp:email(?:\/[A-Za-z0-9@.]+)?)/).replace(/email/g, /[A-Za-z0-9._+-]+@[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(), oa = {
	_backpedal: qr,
	anyPunctuation: Gi,
	anyPunctuationOutsideCode: Ki,
	autolink: qi,
	blockSkip: Mi,
	br: Si,
	code: xi,
	del: qr,
	delLDelim: qr,
	delRDelim: qr,
	emStrongLDelim: Pi,
	emStrongRDelimAst: Ri,
	emStrongRDelimUnd: Vi,
	escape: bi,
	link: Qi,
	nolink: ea,
	punctuation: Di,
	reflink: $i,
	reflinkSearch: ra,
	tag: Yi,
	text: Ci,
	url: qr
}, sa = {
	...oa,
	emStrongLDelim: Ii,
	emStrongRDelimAst: Bi,
	emStrongRDelimUnd: Hi,
	link: J(/^!?\[(label)\]\((.*?)\)/).replace("label", Zi).getRegex(),
	reflink: J(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", Zi).getRegex()
}, ca = {
	...oa,
	emStrongRDelimAst: zi,
	emStrongLDelim: Fi,
	delLDelim: Ui,
	delRDelim: Wi,
	url: J(/^emailProtocol|^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("emailProtocol", aa).replace("protocol", ia).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: J(/^(?:[^a-zA-Z0-9](?=emailProtocol)|(`+|~+|[^`~])(?:(?=[`~])|(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|wwwStart|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9](?=emailProtocol)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))))/).replace("protocol", ia).replace(/emailProtocol/g, /(?:mailto|xmpp):/).replace("wwwStart", Yr ? "(?<![^\\s*_~(])www\\." : "www\\.").getRegex()
}, la = {
	...ca,
	br: J(Si).replace("{2,}", "*").getRegex(),
	text: J(ca.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, ua = {
	normal: gi,
	gfm: vi,
	pedantic: yi
}, da = {
	normal: oa,
	gfm: ca,
	breaks: la,
	pedantic: sa
}, fa = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, pa = (e) => fa[e];
function ma(e, t) {
	if (t) {
		if (Y.escapeTest.test(e)) return e.replace(Y.escapeReplace, pa);
	} else if (Y.escapeTestNoEncode.test(e)) return e.replace(Y.escapeReplaceNoEncode, pa);
	return e;
}
function ha(e) {
	return Y.numericCharacterReference.test(e) ? e.replace(Y.numericCharacterReferenceOrAmpersand, (e, t, n) => {
		if (t === void 0 && n === void 0) return "&amp;";
		let r = t === void 0 ? Number.parseInt(n, 16) : Number.parseInt(t, 10);
		return r === 0 || r > 1114111 || r >= 55296 && r <= 57343 ? "�" : r === 38 ? "&amp;" : String.fromCodePoint(r);
	}) : e;
}
function ga(e) {
	try {
		e = encodeURI(e).replace(Y.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function _a(e, t) {
	let n = e.replace(Y.findPipe, (e, t, n) => {
		let r = !1, i = t;
		for (; --i >= 0 && n[i] === "\\";) r = !r;
		return r ? "|" : " |";
	}).split(Y.splitPipe), r = 0;
	if (n[0].trim() || n.shift(), n.length > 0 && !n.at(-1)?.trim() && n.pop(), t) {
		if (n.length > t) n.splice(t);
		else for (; n.length < t;) n.push("");
	}
	for (; r < n.length; r++) n[r] = n[r].trim().replace(Y.slashPipe, "|");
	return n;
}
function va(e, t, n) {
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
function ya(e) {
	let t = e.split("\n"), n = t.length - 1;
	for (; n >= 0 && Y.blankLine.test(t[n]);) n--;
	return t.length - n <= 2 ? e : t.slice(0, n + 1).join("\n");
}
function ba(e) {
	return e.trim().replace(Y.multipleSpaceGlobal, " ").toLowerCase().toUpperCase().toLowerCase();
}
function xa(e, t) {
	if (e.indexOf(t[0]) === -1 && e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
var Sa = /* @__PURE__ */ new WeakMap();
function Ca(e) {
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
function wa(e, t) {
	let n = Ca(t), r = Sa.get(e);
	r ? r.push(n) : Sa.set(e, [n]);
}
function Ta(e) {
	let t = Sa.get(e);
	t && (t.pop(), t.length === 0 && Sa.delete(e));
}
function Ea(e, t, n) {
	let r = Sa.get(e)?.at(-1);
	return r && r.maskedSrc === t ? r.lastPossibleEmStrongCloserIndex[n] : Ca(t).lastPossibleEmStrongCloserIndex[n];
}
function Da(e, t) {
	let n = Sa.get(e)?.at(-1);
	return n && n.maskedSrc === t ? n.lastPossibleDelCloserIndex : Ca(t).lastPossibleDelCloserIndex;
}
function Oa(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function ka(e, t, n, r, i) {
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
function Aa(e, t, n) {
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
function ja(e, t, n, r) {
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
var Ma = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || Gr;
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
			let e = this.options.pedantic ? t[0] : ya(t[0]);
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
			let e = t[0], n = Aa(e, t[3] || "", this.rules);
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
				let t = va(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceTabChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: va(t[0], "\n"),
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
			raw: va(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = va(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteMarkerWhitespace, (e, t, n) => t + Oa(n, t.length)).replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
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
				let u = t[2].split("\n", 1)[0], d = t[1].length, f = this.options.pedantic ? Oa(u, d) : u.replace(this.rules.other.leadingSpaceTab, (e) => Oa(e, d)), p = e.split("\n", 1)[0], m = !f.trim(), h = 0;
				if (this.options.pedantic ? (h = 2, a = f.trimStart()) : m ? h = d + 1 : (h = f.search(this.rules.other.nonSpaceChar), h = h > 4 ? 1 : h, a = f.slice(h), h += d), m && this.rules.other.blankLine.test(p) && (e = e.substring(p.length + 1), i = !0), !i) {
					let t = this.rules.other.nextBulletRegex(h), n = this.rules.other.hrRegex(h), r = this.rules.other.fencesBeginRegex(h), i = this.options.pedantic ? this.rules.other.pedanticHeadingBeginRegex(h) : this.rules.other.headingBeginRegex(h), o = this.rules.other.htmlBeginRegex(h), s = this.rules.other.blockquoteBeginRegex(h);
					for (; e;) {
						let c = e.split("\n", 1)[0], l;
						if (p = c, this.options.pedantic ? (p = p.replace(this.rules.other.listReplaceNesting, "  "), l = p) : l = p.replace(this.rules.other.leadingSpaceTab, (e) => Oa(e)), r.test(p) || i.test(p) || o.test(p) || s.test(p) || t.test(p) || n.test(p)) break;
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
			let e = ya(t[0]);
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
			if (!this.rules.other.startAngleBracket.test(t[2]) && xa(t[2], "()") !== -1) return;
			let e = ba(t[1]), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: va(t[0], "\n"),
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = _a(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: va(t[0], "\n"),
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
			for (let e of i) a.rows.push(_a(e, a.header.length).map((e, t) => ({
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
				raw: va(t[0], "\n"),
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
			if (!this.options.pedantic && ja(e, t[1], n, this.rules)) return;
			let r = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(r)) {
				if (!this.rules.other.endAngleBracket.test(r)) return;
				let e = va(r.slice(0, -1), "\\");
				if ((r.length - e.length) % 2 == 0) return;
			} else {
				let e = xa(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let i = va(t[0].slice(0, -1), "\\");
			if ((t[0].length - i.length) % 2 == 0) return;
			let a = t[2], o = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(a);
				e && (a = e[1], o = e[3]);
			} else o = t[3] ? t[3].slice(1, -1) : "";
			return a = a.trim(), this.rules.other.startAngleBracket.test(a) && (a = this.options.pedantic && !this.rules.other.endAngleBracket.test(r) ? a.slice(1) : a.slice(1, -1)), ka(t, {
				href: a && a.replace(this.rules.inline.anyPunctuation, "$1"),
				title: o && o.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let r = n[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && ja(e, n[1], r, this.rules)) return;
			let i = t[ba(n[2] || n[1])];
			if (!i) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return ka(n, i, n[0], this.lexer, this.rules);
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
			let _ = i === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd, v = Ea(this, t, i), y = t.length - e.length + n;
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
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim, c = Da(this, t), l = t.length - e.length + n;
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
				text: e ? t[0] : ha(t[0]),
				escaped: e
			};
		}
	}
}, Na = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || Gr, this.options.tokenizer = this.options.tokenizer || new Ma(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			linkEmitted: !1,
			linkParenPossible: !0,
			top: !0
		};
		let t = {
			other: Y,
			block: ua.normal,
			inline: da.normal
		};
		this.options.pedantic ? (t.block = ua.pedantic, t.inline = da.pedantic) : this.options.gfm && (t.block = ua.gfm, t.inline = this.options.breaks ? da.breaks : da.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: ua,
			inline: da
		};
	}
	static lex(t, n) {
		return new e(n).lex(t);
	}
	static lexInline(t, n) {
		return new e(n).inlineTokens(t);
	}
	lex(e) {
		e = e.replace(Y.carriageReturn, "\n"), this.blockTokens(e, this.tokens);
		for (let e = 0; e < this.inlineQueue.length; e++) {
			let t = this.inlineQueue[e];
			this.state.inRawBlock = !1, this.inlineTokens(t.src, t.tokens);
		}
		return this.inlineQueue = [], this.tokens;
	}
	blockTokens(e, t = [], n = !1) {
		this.tokenizer.lexer = this, this.options.pedantic && (e = e.replace(Y.tabCharGlobal, "    ").replace(Y.spaceLine, ""));
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
			if (e.charAt(0) !== "!" && Object.hasOwn(this.tokens.links, ba(e.slice(n + 1, -1))) && !(n > 1 && this.linkInText(e.slice(1, n - 1)))) return !0;
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
		let n = e.includes("*") || e.includes("_") || e.includes("~") || this.options.hooks?.emStrongMask || this.tokenizer.emStrong !== Ma.prototype.emStrong || this.tokenizer.del !== Ma.prototype.del ? this.#n(e) : e;
		wa(this.tokenizer, n);
		try {
			return this.#t(e, n, t);
		} finally {
			Ta(this.tokenizer);
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
				if (!Object.hasOwn(this.tokens.links, ba(t.slice(r + 1, -1)))) return t;
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
}, Pa = class {
	options;
	parser;
	constructor(e) {
		this.options = e || Gr;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(Y.notSpaceStart)?.[0], i = e ? e.replace(Y.endingNewline, "") + "\n" : "";
		return r ? "<pre><code class=\"language-" + ma(r) + "\">" + (n ? i : ma(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : ma(i, !0)) + "</code></pre>\n";
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
		return `<code>${ma(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, text: n, tokens: r, autolink: i }) {
		let a = i ? ma(n, !0) : this.parser.parseInline(r), o = ga(e);
		if (o === null) return a;
		e = ma(o, i);
		let s = "<a href=\"" + e + "\"";
		return t && (s += " title=\"" + ma(t) + "\""), s += ">" + a + "</a>", s;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = ga(e);
		if (i === null) return ma(n);
		e = i;
		let a = `<img src="${ma(e)}" alt="${ma(n)}"`;
		return t && (a += ` title="${ma(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : ma(e.text);
	}
}, Fa = class {
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
}, Ia = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || Gr, this.options.renderer = this.options.renderer || new Pa(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new Fa(), this.textRenderer.parser = this;
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
}, La = class {
	options;
	block;
	constructor(e) {
		this.options = e || Gr;
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
		return e ? Na.lex : Na.lexInline;
	}
	provideParser(e = this.block) {
		return e ? Ia.parse : Ia.parseInline;
	}
};
function Ra(e, t) {
	if (Array.isArray(t)) for (let n = 0; n < t.length; n++) e.push(t[n]);
	else e.push(t);
}
var za = new class {
	defaults = Wr();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = Ia;
	Renderer = Pa;
	TextRenderer = Fa;
	Lexer = Na;
	Tokenizer = Ma;
	Hooks = La;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (Ra(n, t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) Ra(n, this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) Ra(n, this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				Ra(n, this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					Ra(n, this.walkTokens(i, t));
				}) : e.tokens && Ra(n, this.walkTokens(e.tokens, t));
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
				let t = this.defaults.renderer || new Pa(this.defaults);
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
				let t = this.defaults.tokenizer || new Ma(this.defaults);
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
				let t = this.defaults.hooks || new La();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = La.passThroughHooks.has(n) ? (e) => {
						if (t.options.async && La.passThroughHooksRespectAsync.has(n)) return (async () => {
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
					return Ra(n, r.call(this, e)), t && Ra(n, t.call(this, e)), n;
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
		return Na.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return Ia.parse(e, t ?? this.defaults);
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
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? Na.lex : Na.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? Ia.parse : Ia.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? Na.lex : Na.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? Ia.parse : Ia.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + ma(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function X(e, t) {
	return za.parse(e, t);
}
X.options = X.setOptions = function(e) {
	return za.setOptions(e), X.defaults = za.defaults, Kr(X.defaults), X;
}, X.getDefaults = Wr, X.defaults = Gr;
function Ba(...e) {
	return za.use(...e), X.defaults = za.defaults, Kr(X.defaults), X;
}
X.use = Ba, X.walkTokens = function(e, t) {
	return za.walkTokens(e, t);
}, X.parseInline = za.parseInline, X.Parser = Ia, X.parser = Ia.parse, X.Renderer = Pa, X.TextRenderer = Fa, X.Lexer = Na, X.lexer = Na.lex, X.Tokenizer = Ma, X.Hooks = La, X.parse = X, X.options, X.setOptions, X.walkTokens, X.parseInline, Ia.parse, Na.lex;
// node_modules/dompurify/dist/purify.es.mjs
function Va(e, t) {
	this.v = e, this.k = t;
}
function Ha(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function Ua(e) {
	if (Array.isArray(e)) return e;
}
function Wa(e, t) {
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
function Ga() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function Ka(e, t) {
	return Ua(e) || Wa(e, t) || qa(e, t) || Ga();
}
function qa(e, t) {
	if (e) {
		if (typeof e == "string") return Ha(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? Ha(e, t) : void 0;
	}
}
function Ja(e) {
	var t, n;
	function r(t, n) {
		try {
			var a = e[t](n), o = a.value, s = o instanceof Va;
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
Ja.prototype[typeof Symbol == "function" && Symbol.asyncIterator || "@@asyncIterator"] = function() {
	return this;
}, Ja.prototype.next = function(e) {
	return this._invoke("next", e);
}, Ja.prototype.throw = function(e) {
	return this._invoke("throw", e);
}, Ja.prototype.return = function(e) {
	return this._invoke("return", e);
};
var Ya = Object.entries, Xa = Object.setPrototypeOf, Za = Object.isFrozen, Qa = Object.getPrototypeOf, $a = Object.getOwnPropertyDescriptor, eo = Object.freeze, to = Object.seal, no = Object.create, ro = typeof Reflect < "u" && Reflect, io = ro.apply, ao = ro.construct;
eo ||= function(e) {
	return e;
}, to ||= function(e) {
	return e;
}, io ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, ao ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var oo = Z(Array.prototype.forEach);
Array.prototype.indexOf;
var so = Z(Array.prototype.lastIndexOf), co = Z(Array.prototype.pop), lo = Z(Array.prototype.push), uo = Z(Array.prototype.slice), fo = Z(Array.prototype.splice), po = Array.isArray, mo = jo(Set), ho = Z(Set.prototype.add), go = Z(Set.prototype.has), _o = Z(String.prototype.toLowerCase), vo = Z(String.prototype.toString), yo = Z(String.prototype.match), bo = Z(String.prototype.replace), xo = Z(String.prototype.indexOf), So = Z(String.prototype.trim), Co = Z(Number.prototype.toString), wo = Z(Boolean.prototype.toString), To = typeof BigInt > "u" ? null : Z(BigInt.prototype.toString), Eo = typeof Symbol > "u" ? null : Z(Symbol.prototype.toString), Do = Z(Object.prototype.hasOwnProperty), Oo = Z(Object.prototype.toString), ko = Z(RegExp.prototype.test), Ao = jo(TypeError);
function Z(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return io(e, t, n);
	};
}
function jo(e) {
	return function() {
		return ao(e, [...arguments]);
	};
}
function Q(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : _o;
	if (Xa && Xa(e, null), !po(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (Za(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function Mo(e) {
	for (let t = 0; t < e.length; t++) Do(e, t) || (e[t] = null);
	return e;
}
function No(e) {
	let t = no(null);
	for (let r of Ya(e)) {
		var n = Ka(r, 2);
		let i = n[0], a = n[1];
		Do(e, i) && (t[i] = po(a) ? Mo(a) : a && typeof a == "object" && a.constructor === Object ? No(a) : a);
	}
	return t;
}
function Po(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return Co(e);
		case "boolean": return wo(e);
		case "bigint": return To ? To(e) : "0";
		case "symbol": return Eo ? Eo(e) : "Symbol()";
		case "undefined": return Oo(e);
		case "function":
		case "object": {
			if (e === null) return Oo(e);
			let t = e, n = Fo(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : Oo(e);
			}
			return Oo(e);
		}
		default: return Oo(e);
	}
}
function Fo(e, t) {
	for (; e !== null;) {
		let n = $a(e, t);
		if (n) {
			if (n.get) return Z(n.get);
			if (typeof n.value == "function") return Z(n.value);
		}
		e = Qa(e);
	}
	function n() {
		return null;
	}
	return n;
}
function Io(e, t) {
	for (; e !== null;) {
		let n = $a(e, t);
		if (n) return n.get ? Z(n.get) : null;
		e = Qa(e);
	}
	return null;
}
function Lo(e) {
	try {
		return ko(e, ""), !0;
	} catch {
		return !1;
	}
}
var Ro = eo(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), zo = eo(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), Bo = eo([
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
]), Vo = eo([
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
]), Ho = eo(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), Uo = eo([
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
]), Wo = eo(["#text"]), Go = eo(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), Ko = eo(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dominant-baseline.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.pointer-events.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-orientation.text-rendering.textlength.type.u1.u2.unicode.values.vector-effect.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), qo = eo(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), Jo = eo([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), Yo = to(/{{[\w\W]*|^[\w\W]*}}/g), Xo = to(/<%[\w\W]*|^[\w\W]*%>/g), Zo = to(/\${[\w\W]*/g), Qo = to(/^data-[\-\w.\u00B7-\uFFFF]+$/), $o = to(/^aria-[\-\w]+$/), es = to(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), ts = to(/^(?:\w+script|data):/i), ns = to(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), rs = to(/^html$/i), is = to(/^[a-z][.\w]*(-[.\w]+)+$/i), as = to(/<[/\w!]/g), os = to(/<[/\w]/g), ss = to(/<\/no(script|embed|frames)/i), cs = to(/\/>/i), $ = {
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
}, ls = eo(Q({}, [
	"style",
	"script",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"plaintext",
	"noscript"
])), us = function() {
	return typeof window > "u" ? null : window;
}, ds = function(e, t) {
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
}, fs = function() {
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
}, ps = function(e, t, n, r) {
	return Do(e, t) && po(e[t]) ? Q(r.base ? No(r.base) : {}, e[t], r.transform) : n;
}, ms = function(e, t, n) {
	let r = Do(e, t) ? e[t] : void 0;
	return r && typeof r == "object" ? No(r) : n();
};
function hs() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : us(), t = (e) => hs(e);
	if (t.version = "3.4.17", t.removed = [], !e || !e.document || e.document.nodeType !== $.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = Fo(d, "cloneNode"), p = Fo(d, "remove"), m = Fo(d, "removeAttributeNode"), h = Fo(d, "nextSibling"), g = Fo(d, "childNodes"), _ = Fo(d, "parentNode"), v = Fo(d, "shadowRoot"), y = Fo(d, "attributes"), b = o && o.prototype ? Fo(o.prototype, "nodeType") : null, x = o && o.prototype ? Fo(o.prototype, "nodeName") : null, S = o && o.prototype ? Fo(o.prototype, "ownerDocument") : null, C = function(e) {
		return b ? b(e) : e.nodeType;
	}, w = function(e) {
		return x ? x(e) : e.nodeName;
	}, ee = Io(d, "namespaceURI"), te = function(e) {
		return ee ? ee(e) : e.namespaceURI;
	};
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let T, ne = "", re, ie = !1, ae = 0, oe = function() {
		if (ae > 0) throw Ao("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
	}, se = function(e) {
		oe(), ae++;
		try {
			return T.createHTML(e);
		} finally {
			ae--;
		}
	}, ce = function(e) {
		oe(), ae++;
		try {
			return T.createScriptURL(e);
		} finally {
			ae--;
		}
	}, le = function() {
		return ie ||= (re = ds(u, i), !0), re;
	}, ue = n, de = ue.implementation, fe = ue.createNodeIterator, pe = ue.createDocumentFragment, me = ue.getElementsByTagName, he = r.importNode, E = fs();
	t.isSupported = typeof Ya == "function" && typeof _ == "function" && de && de.createHTMLDocument !== void 0;
	let ge = Yo, D = Xo, _e = Zo, O = Qo, ve = $o, ye = ts, k = ns, be = is, xe = es, A = null, Se = Q({}, [
		...Ro,
		...zo,
		...Bo,
		...Ho,
		...Wo
	]), j = null, Ce = Q({}, [
		...Go,
		...Ko,
		...qo,
		...Jo
	]), we = Object.seal(no(null, {
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
	})), Te = null, Ee = null, De = Object.seal(no(null, {
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
	})), Oe = !0, ke = !0, Ae = !1, je = !0, Me = !1, Ne = !0, Pe = !1, Fe = !1, Ie = null, Le = null, Re = !1, ze = !1, Be = !1, Ve = !1, He = !0, Ue = !1, We = "user-content-", M = !0, Ge = !1, Ke = null, qe = {}, Je = null, Ye = Q({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Xe = null, Ze = Q({}, [
		"audio",
		"video",
		"img",
		"source",
		"image",
		"track"
	]), Qe = null, $e = Q({}, [
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
	]), et = "http://www.w3.org/1998/Math/MathML", tt = "http://www.w3.org/2000/svg", nt = "http://www.w3.org/1999/xhtml", rt = nt, it = !1, at = null, ot = Q({}, [
		et,
		tt,
		nt
	], vo), st = eo([
		"mi",
		"mo",
		"mn",
		"ms",
		"mtext"
	]), ct = Q({}, st), N = eo(["annotation-xml"]), lt = Q({}, N), ut = Q({}, [
		"title",
		"style",
		"font",
		"a",
		"script"
	]), dt = null, ft = ["application/xhtml+xml", "text/html"], P = null, pt = null, mt = n.createElement("form"), ht = function(e) {
		return e instanceof RegExp || e instanceof Function;
	}, gt = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		if (pt && pt === e) return;
		e && typeof e == "object" || (e = {}), e = No(e), dt = ft.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, P = dt === "application/xhtml+xml" ? vo : _o, A = ps(e, "ALLOWED_TAGS", Se, { transform: P }), j = ps(e, "ALLOWED_ATTR", Ce, { transform: P }), at = ps(e, "ALLOWED_NAMESPACES", ot, { transform: vo }), Qe = ps(e, "ADD_URI_SAFE_ATTR", $e, {
			transform: P,
			base: $e
		}), Xe = ps(e, "ADD_DATA_URI_TAGS", Ze, {
			transform: P,
			base: Ze
		}), Je = ps(e, "FORBID_CONTENTS", Ye, { transform: P }), Te = ps(e, "FORBID_TAGS", No({}), { transform: P }), Ee = ps(e, "FORBID_ATTR", No({}), { transform: P }), qe = Do(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? No(e.USE_PROFILES) : e.USE_PROFILES : !1, Oe = e.ALLOW_ARIA_ATTR !== !1, ke = e.ALLOW_DATA_ATTR !== !1, Ae = e.ALLOW_UNKNOWN_PROTOCOLS || !1, je = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, Me = e.SAFE_FOR_TEMPLATES || !1, Ne = e.SAFE_FOR_XML !== !1, Pe = e.WHOLE_DOCUMENT || !1, ze = e.RETURN_DOM || !1, Be = e.RETURN_DOM_FRAGMENT || !1, Ve = e.RETURN_TRUSTED_TYPE || !1, Re = e.FORCE_BODY || !1, He = e.SANITIZE_DOM !== !1, Ue = e.SANITIZE_NAMED_PROPS || !1, M = e.KEEP_CONTENT !== !1, Ge = e.IN_PLACE || !1, xe = Lo(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : es, rt = typeof e.NAMESPACE == "string" ? e.NAMESPACE : nt, ct = ms(e, "MATHML_TEXT_INTEGRATION_POINTS", () => Q({}, st)), lt = ms(e, "HTML_INTEGRATION_POINTS", () => Q({}, N));
		let t = ms(e, "CUSTOM_ELEMENT_HANDLING", () => no(null));
		if (we = no(null), Do(t, "tagNameCheck") && ht(t.tagNameCheck) && (we.tagNameCheck = t.tagNameCheck), Do(t, "attributeNameCheck") && ht(t.attributeNameCheck) && (we.attributeNameCheck = t.attributeNameCheck), Do(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (we.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), to(we), Me && (ke = !1), Be && (ze = !0), qe && (A = Q({}, Wo), j = no(null), qe.html === !0 && (Q(A, Ro), Q(j, Go)), qe.svg === !0 && (Q(A, zo), Q(j, Ko), Q(j, Jo)), qe.svgFilters === !0 && (Q(A, Bo), Q(j, Ko), Q(j, Jo)), qe.mathMl === !0 && (Q(A, Ho), Q(j, qo), Q(j, Jo))), De.tagCheck = null, De.attributeCheck = null, Do(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? De.tagCheck = e.ADD_TAGS : po(e.ADD_TAGS) && (A === Se && (A = No(A)), Q(A, e.ADD_TAGS, P))), Do(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? De.attributeCheck = e.ADD_ATTR : po(e.ADD_ATTR) && (j === Ce && (j = No(j)), Q(j, e.ADD_ATTR, P))), Do(e, "ADD_FORBID_CONTENTS") && po(e.ADD_FORBID_CONTENTS) && (Je === Ye && (Je = No(Je)), Q(Je, e.ADD_FORBID_CONTENTS, P)), M && (A["#text"] = !0), Pe && Q(A, [
			"html",
			"head",
			"body"
		]), A.table && (Q(A, ["tbody"]), delete Te.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw Ao("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw Ao("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
			let t = T;
			T = e.TRUSTED_TYPES_POLICY;
			try {
				ne = se("");
			} catch (e) {
				throw T = t, e;
			}
		} else e.TRUSTED_TYPES_POLICY === null ? (T = void 0, ne = "") : (T === void 0 && (T = le()), T && typeof ne == "string" && (ne = se("")));
		eo && eo(e), pt = e;
	}, _t = Q({}, [
		...zo,
		...Bo,
		...Vo
	]), vt = Q({}, [...Ho, ...Uo]), yt = function(e, t, n) {
		return t.namespaceURI === nt ? e === "svg" : t.namespaceURI === et ? e === "svg" && (n === "annotation-xml" || ct[n]) : !!_t[e];
	}, bt = function(e, t, n) {
		return t.namespaceURI === nt ? e === "math" : t.namespaceURI === tt ? e === "math" && lt[n] : !!vt[e];
	}, xt = function(e, t, n) {
		return t.namespaceURI === tt && !lt[n] || t.namespaceURI === et && !ct[n] ? !1 : !vt[e] && (ut[e] || !_t[e]);
	}, St = function(e) {
		let t = te(e), n = _o(w(e)), r = _(e), i, a;
		return r && C(r) === $.element ? (i = { namespaceURI: te(r) }, a = _o(w(r))) : (i = { namespaceURI: rt }, a = "template"), at[t] ? t === tt ? yt(n, i, a) : t === et ? bt(n, i, a) : t === nt ? xt(n, i, a) : !!(dt === "application/xhtml+xml" && at[t]) : !1;
	}, Ct = function(e) {
		lo(t.removed, { element: e });
		try {
			_(e).removeChild(e);
		} catch {
			if (p(e), !_(e)) throw Ao("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
		}
	}, wt = function(e, t, n) {
		try {
			m(e, t);
		} catch {
			try {
				e.removeAttribute(n);
			} catch {}
		}
	}, Tt = function(e) {
		Ot(e);
		let t = g(e);
		t && oo(uo(t), (e) => {
			try {
				p(e);
			} catch {}
		});
		let n = C(e) === $.element ? y(e) : null;
		if (n) for (let t = n.length - 1; t >= 0; --t) {
			let r = n[t], i = r && r.name;
			typeof i == "string" && wt(e, r, i);
		}
	}, Et = function(e, n, r) {
		if (!r) try {
			r = n.getAttributeNode(e);
		} catch {
			r = null;
		}
		lo(t.removed, {
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
			if (ze || Be) try {
				Ct(n);
			} catch {}
			else try {
				n.setAttribute(e, "");
			} catch {}
		}
	}, Dt = function(e) {
		let t = y(e);
		if (t) for (let n = t.length - 1; n >= 0; --n) {
			let r = t[n], i = r && r.name;
			typeof i != "string" || j[P(i)] || wt(e, r, i);
		}
	}, F = function(e, t) {
		let n = arguments.length > 2 && arguments[2] !== void 0 && arguments[2], r = [e];
		for (; r.length > 0;) {
			let e = r.pop();
			if (t(e) === !1) continue;
			if (n && C(e) === $.element) {
				let t = v(e);
				Vt(t) && r.push(t);
			}
			let i = g(e);
			if (i) for (let e = i.length - 1; e >= 0; --e) r.push(i[e]);
		}
	}, Ot = function(e) {
		F(e, (e) => {
			C(e) === $.element && Dt(e);
		});
	}, I = function(e) {
		let t = [];
		return F(e, (e) => {
			t.push(e);
		}, !0), t;
	}, kt = function(e, t) {
		let n = mo();
		oo(I(e), (e) => {
			C(e) === $.element && ho(n, e);
		}), oo(t, (e) => {
			C(e) === $.element && !go(n, e) && Dt(e);
		});
	}, At = function(e, n) {
		let r = !1;
		return oo(t.removed, (t) => {
			t.element && (r ||= t.element === e, Ot(t.element));
		}), kt(e, n), r;
	}, jt = function(e, t) {
		let n = Ke, r = null, i = () => (r || (r = mo(), oo(t, (e) => {
			ho(r, e);
		})), r);
		for (let t = 0;; t++) {
			let r = I(e), a = [e], o = !1;
			if (oo(r, (e) => {
				C(e) === $.documentFragment ? lo(a, e) : !o && !go(n, e) && (o = go(i(), e));
			}), !o) return;
			if (t >= 8) throw Ao("the tree kept changing during in-place sanitization; refusing to return it");
			oo(a, (t, n) => {
				Mt(t, n > 0, e, i());
			});
		}
	}, Mt = function(e, t, n, r) {
		let i = Lt(e), a;
		for (; a = i.nextNode();) go(r, a) && !go(Ke, a) && (t && R(E.uponSanitizeShadowNode, a, null), Kt(a, n), $t(a, n), Vt(a.content) && en(a.content));
	}, Nt = function(e) {
		Ke && ho(Ke, e);
	}, Pt = function(e, t) {
		return Ne ? e === "patchsrc" || e === "for" && t !== "label" && t !== "output" : !1;
	}, Ft = function(e) {
		Ne && F(e, (e) => {
			let t = C(e);
			if (t === $.processingInstruction || t === $.comment && ko(os, e.data)) {
				try {
					p(e);
				} catch {}
				return !1;
			}
			if (t === $.element) {
				let t = e, n = P(w(e));
				try {
					t.hasAttribute && t.hasAttribute("patchsrc") && t.removeAttribute("patchsrc"), t.hasAttribute && t.hasAttribute("for") && Pt("for", n) && t.removeAttribute("for");
				} catch {}
			}
			return !0;
		});
	}, It = function(e) {
		let t = null, r = null;
		if (Re) e = "<remove></remove>" + e;
		else {
			let t = yo(e, /^[\r\n\t ]+/);
			r = t && t[0];
		}
		dt === "application/xhtml+xml" && rt === nt && (e = "<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>" + e + "</body></html>");
		let i = T ? se(e) : e;
		if (rt === nt) try {
			t = new l().parseFromString(i, dt);
		} catch {}
		if (!t || !t.documentElement) {
			t = de.createDocument(rt, "template", null);
			try {
				t.documentElement.innerHTML = it ? ne : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), rt === nt ? me.call(t, Pe ? "html" : "body")[0] : Pe ? t.documentElement : a;
	}, Lt = function(e) {
		let t = S ? S(e) : e.ownerDocument;
		return fe.call(t || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, Rt = function(e) {
		return e = bo(e, ge, " "), e = bo(e, D, " "), e = bo(e, _e, " "), e;
	}, zt = function(e) {
		Bt(e);
		let t = e.querySelectorAll?.call(e, "template");
		t && oo(t, (e) => {
			Vt(e.content) && zt(e.content);
		});
	}, Bt = function(e) {
		e.normalize();
		let t = S ? S(e) : e.ownerDocument, n = fe.call(t || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), r = n.nextNode();
		for (; r;) r.data = Rt(r.data), r = n.nextNode();
	}, L = function(e) {
		let t = x ? x(e) : null;
		return typeof t != "string" || P(t) !== "form" ? !1 : typeof e.nodeName != "string" || typeof e.textContent != "string" || typeof e.removeChild != "function" || e.attributes !== y(e) || typeof e.removeAttribute != "function" || typeof e.removeAttributeNode != "function" || typeof e.getAttributeNode != "function" || typeof e.setAttribute != "function" || typeof e.namespaceURI != "string" || typeof e.insertBefore != "function" || typeof e.hasChildNodes != "function" || e.nodeType !== b(e) || e.childNodes !== g(e);
	}, Vt = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return b(e) === $.documentFragment;
		} catch {
			return !1;
		}
	}, Ht = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return typeof b(e) == "number";
		} catch {
			return !1;
		}
	};
	function R(e, n, r) {
		e.length !== 0 && oo(e, (e) => {
			e.call(t, n, r, pt);
		});
	}
	let Ut = function(e, t) {
		return !!(Ne && e.hasChildNodes() && !Ht(e.firstElementChild) && ko(as, e.textContent) && ko(as, e.innerHTML) || Ne && ls[t] && C(e) === $.element && te(e) === nt && (Ht(e.firstElementChild) || typeof e.textContent == "string" && ko(as, e.textContent)) || e.nodeType === $.processingInstruction || Ne && e.nodeType === $.comment && ko(os, e.data));
	}, z = function(e, t) {
		return e instanceof RegExp ? ko(e, t) : e instanceof Function && !!e(t, ...[...arguments].slice(2));
	}, B = function(e, t, n) {
		if (!Te[t] && Yt(t) && z(we.tagNameCheck, t)) return !1;
		if (M && !Je[t] && !Je[_o(t)]) {
			let t = _(e), r = g(e);
			if (r && t) {
				let i = r.length;
				for (let a = i - 1; a >= 0; --a) {
					let i = e === n ? f(r[a], !0) : r[a];
					t.insertBefore(i, h(e));
				}
			}
		}
		return Ct(e), !0;
	}, Wt = function(e, t, n, r) {
		return e.length === 0 ? t : t === n || t === r ? No(t) : t;
	}, Gt = function(e, t) {
		return e === t || _(e) !== null ? !1 : (Ge && Ot(e), !0);
	}, Kt = function(e, n) {
		if (Nt(e), R(E.beforeSanitizeElements, e, null), Gt(e, n)) return !0;
		if (L(e)) return Ct(e), !0;
		let r = P(w(e));
		if (A = Wt(E.uponSanitizeElement, A, Se, Ie), R(E.uponSanitizeElement, e, {
			tagName: r,
			allowedTags: A
		}), Gt(e, n)) return !0;
		if (Ut(e, r)) return Ct(e), !0;
		if (Te[r] || !(De.tagCheck instanceof Function && De.tagCheck(r)) && !A[r]) {
			let t = B(e, r, n);
			return t === !1 && (R(E.afterSanitizeElements, e, null), Gt(e, n)) ? !0 : t;
		}
		if (C(e) === $.element && !St(e) || (r === "noscript" || r === "noembed" || r === "noframes") && ko(ss, e.innerHTML)) return Ct(e), !0;
		if (Me && e.nodeType === $.text) {
			let n = Rt(e.textContent);
			e.textContent !== n && (lo(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return R(E.afterSanitizeElements, e, null), Gt(e, n);
	}, qt = function(e, t, r) {
		if (Ee[t] || Pt(t, e) || He && (t === "id" || t === "name") && (r in n || r in mt)) return !1;
		let i = j[t] || De.attributeCheck instanceof Function && De.attributeCheck(t, e);
		return ke && ko(O, t) || Oe && ko(ve, t) ? !0 : i ? Qe[t] || ko(xe, bo(r, k, "")) || (t === "src" || t === "xlink:href" || t === "href") && e !== "script" && xo(r, "data:") === 0 && Xe[e] || Ae && !ko(ye, bo(r, k, "")) ? !0 : !r : Yt(e) && z(we.tagNameCheck, e) && z(we.attributeNameCheck, t, e) || t === "is" && we.allowCustomizedBuiltInElements && z(we.tagNameCheck, r);
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
		return !Jt[_o(e)] && ko(be, e);
	}, Xt = function(e, t, n, r) {
		if (T && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return se(r);
			case "TrustedScriptURL": return ce(r);
		}
		return r;
	}, Zt = function(e, t, n, r) {
		try {
			return n ? e.setAttributeNS(n, t, r) : e.setAttribute(t, r), !L(e) || (Ct(e), !1);
		} catch {
			return Et(t, e), !1;
		}
	}, Qt = function(e, t) {
		return Ne && ko(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, t) ? !0 : _o(e) === "attributename" && yo(t, "href") !== null;
	}, $t = function(e, n) {
		if (R(E.beforeSanitizeAttributes, e, null), Gt(e, n)) return;
		let r = e.attributes;
		if (!r || L(e)) return;
		j = Wt(E.uponSanitizeAttribute, j, Ce, Le);
		let i = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: j,
			forceKeepAttr: void 0
		}, a = r.length, o = P(e.nodeName);
		for (; a--;) {
			let n = r[a], s = n.name, c = n.namespaceURI, l = n.value, u = P(s), d = l, f = s === "value" ? d : So(d), p = !1;
			i.attrName = u, i.attrValue = f, i.keepAttr = !0, i.forceKeepAttr = void 0, R(E.uponSanitizeAttribute, e, i), f = i.attrValue, Ue && (u === "id" || u === "name") && xo(f, We) !== 0 && (Et(s, e, n), f = We + f, p = !0), Qt(s, f) ? Et(s, e, n) : i.forceKeepAttr || (i.keepAttr ? !je && ko(cs, f) ? Et(s, e, n) : (Me && (f = Rt(f)), qt(o, u, f) ? (f = Xt(o, u, c, f), f !== d && Zt(e, s, c, f) && p && co(t.removed)) : Et(s, e, n)) : Et(s, e, n));
		}
		R(E.afterSanitizeAttributes, e, null), Gt(e, n);
	}, en = function(e) {
		let t = null, n = Lt(e);
		for (R(E.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (R(E.uponSanitizeShadowNode, t, null), Kt(t, e), $t(t, e), Vt(t.content) && en(t.content), C(t) === $.element) {
			let e = v(t);
			Vt(e) && (tn(e), en(e));
		}
		Me && Bt(e), R(E.afterSanitizeShadowDOM, e, null);
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
				if (typeof e == "string" && P(e) === "template") {
					let e = n.content;
					Vt(e) && t.push({
						node: e,
						shadow: null
					});
				}
			}
			if (r) {
				let e = v(n);
				Vt(e) && t.push({
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
		if (it = !e, it && (e = "<!-->"), typeof e != "string" && !Ht(e) && (e = Po(e), typeof e != "string")) throw Ao("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		Fe ? (A = Ie, j = Le) : gt(n), (E.uponSanitizeElement.length > 0 || E.uponSanitizeAttribute.length > 0) && (A = No(A)), E.uponSanitizeAttribute.length > 0 && (j = No(j)), t.removed = [];
		let c = Ge && typeof e != "string" && Ht(e), l = [], u = Ke;
		if (c) {
			Ft(e);
			let t = w(e);
			if (typeof t == "string") {
				let n = P(t);
				if (!A[n] || Te[n]) throw Tt(e), Ao("root node is forbidden and cannot be sanitized in-place");
			}
			if (L(e)) throw Tt(e), Ao("root node is clobbered and cannot be sanitized in-place");
			l = I(e), Ke = mo();
		} else if (Ht(e)) i = It("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === $.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), tn(i);
		else {
			if (!ze && !Me && !Pe && e.indexOf("<") === -1) return T && Ve ? se(e) : e;
			if (i = It(e), !i) return ze ? null : Ve ? ne : "";
		}
		i && Re && Ct(i.firstChild);
		let d = c ? e : i;
		try {
			c && tn(e);
			let t = Lt(d);
			for (; o = t.nextNode();) Kt(o, d), $t(o, d), Vt(o.content) && en(o.content);
			c && jt(e, l);
		} catch (t) {
			throw c && (Tt(e), At(e, l)), t;
		} finally {
			Ke = u;
		}
		if (c) {
			if (At(e, l)) throw Ao("a node selected for removal could not be safely returned; refusing to sanitize in place");
			return Me && zt(e), e;
		}
		if (ze) {
			if (Me && zt(i), Be) for (s = pe.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (j.shadowroot || j.shadowrootmode) && (s = he.call(r, s, !0)), s;
		}
		let f = Pe ? i.outerHTML : i.innerHTML;
		return Pe && A["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && ko(rs, i.ownerDocument.doctype.name) && (f = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + f), Me && (f = Rt(f)), T && Ve ? se(f) : f;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		gt(e), Fe = !0, Ie = A, Le = j;
	}, t.clearConfig = function() {
		pt = null, Fe = !1, Ie = null, Le = null, T = re, ne = "";
	}, t.isValidAttribute = function(e, t, n) {
		pt || gt({});
		let r = P(e), i = P(t);
		return typeof n == "string" && (Qt(t, n) || !je && ko(cs, n)) ? !1 : qt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && Do(E, e) && lo(E[e], t);
	}, t.removeHook = function(e, t) {
		if (Do(E, e)) {
			if (t !== void 0) {
				let n = so(E[e], t);
				return n === -1 ? void 0 : fo(E[e], n, 1)[0];
			}
			return co(E[e]);
		}
	}, t.removeHooks = function(e) {
		Do(E, e) && (E[e] = []);
	}, t.removeAllHooks = function() {
		E = fs();
	}, t;
}
hs();
// kit/format.js
function gs(e) {
	try {
		let t = new URL(e);
		return t.protocol === "https:" || t.protocol === "http:";
	} catch {
		return !1;
	}
}
// kit/Holder.svelte
var _s = /* @__PURE__ */ G("<span class=\"svelte-bwb1ob\"> </span>"), vs = /* @__PURE__ */ G("<span class=\"faint svelte-bwb1ob\" data-testid=\"holder-pool\"> </span>"), ys = /* @__PURE__ */ G("<span class=\"tool muted svelte-bwb1ob\" data-testid=\"holder-tool\"> </span>"), bs = /* @__PURE__ */ G("<button type=\"button\" class=\"holder svelte-bwb1ob\" data-testid=\"holder\"><span></span> <!> <span class=\"name svelte-bwb1ob\" data-testid=\"holder-name\"> </span> <!> <!></button>");
function xs(e, t) {
	Le(t, !0);
	let n = Vr(t, "label", 3, "");
	var r = bs(), i = R(r);
	let a;
	var o = B(i, 2), s = (e) => {
		var t = _s(), r = z(t, !0);
		V(() => q(r, n())), K(e, t);
	};
	yr(o, (e) => {
		n() && e(s);
	});
	var c = B(o, 2), l = z(c, !0), u = B(c, 2), d = (e) => {
		var n = vs(), r = z(n, !0);
		V((e) => q(r, e), [() => [t.h.pool, t.h.model].filter(Boolean).join(" · ")]), K(e, n);
	};
	yr(u, (e) => {
		(t.h.pool || t.h.model) && e(d);
	});
	var f = B(u, 2), p = (e) => {
		var n = ys(), r = z(n, !0);
		V(() => q(r, t.h.tool || "working")), K(e, n);
	};
	yr(f, (e) => {
		t.h.busy && e(p);
	}), k(r), V(() => {
		Fr(r, "title", "Open " + t.h.id + " read-only"), Fr(r, "data-session", t.h.id), a = jr(i, 1, "dot svelte-bwb1ob", null, a, {
			busy: t.h.busy,
			dead: t.h.alive === !1
		}), q(l, t.h.display_name || t.h.id);
	}), Xn("click", r, (e) => {
		e.stopPropagation(), t.onopen(t.h.id);
	}), K(e, r), Re();
}
Zn(["click"]);
// plugins/attention/Attention.svelte
var Ss = /* @__PURE__ */ G("<button class=\"ghost\" data-testid=\"push-test\">Test</button>"), Cs = /* @__PURE__ */ G("<!> <button data-testid=\"push-toggle\"> </button>", 1), ws = /* @__PURE__ */ G("<span class=\"faint pushna svelte-t4vask\" data-testid=\"push-state\"> </span>"), Ts = /* @__PURE__ */ G("<div class=\"card err\"> </div>"), Es = /* @__PURE__ */ G("<p class=\"faint\">Loading…</p>"), Ds = /* @__PURE__ */ G("<div class=\"card empty\" data-testid=\"attention-empty\"><div class=\"big\">Nothing needs you.</div> <div class=\"muted\">No escalations, questions, decisions or other items waiting on you.</div></div>"), Os = /* @__PURE__ */ G("<a target=\"_blank\" rel=\"noopener noreferrer\" data-testid=\"attention-ref\"> </a>"), ks = /* @__PURE__ */ G("· <!>", 1), As = /* @__PURE__ */ G("<div class=\"summary muted svelte-t4vask\" data-testid=\"attention-summary\"> </div>"), js = /* @__PURE__ */ G("<button type=\"button\" data-testid=\"row-resolve\" class=\"svelte-t4vask\"> </button>"), Ms = /* @__PURE__ */ G("<tr data-testid=\"attention-row\"><td class=\"c-id svelte-t4vask\"><span class=\"pill warn\"> </span></td><td class=\"c-title svelte-t4vask\"> <div class=\"faint mono\"> <!></div> <!> <div class=\"acts svelte-t4vask\" role=\"presentation\"><!> <button type=\"button\" class=\"ghost svelte-t4vask\" data-testid=\"open-chat\">Open chat</button></div></td><td class=\"muted full svelte-t4vask\"> </td><td class=\"muted c-time svelte-t4vask\"> </td></tr>"), Ns = /* @__PURE__ */ G("<div class=\"card scroll\"><table class=\"stack\" data-testid=\"attention-list\"><thead><tr><th>Kind</th><th>Item</th><th>Why</th><th>Since</th></tr></thead><tbody></tbody></table></div>"), Ps = /* @__PURE__ */ G("<span class=\"step muted svelte-t4vask\"> </span> <!>", 1), Fs = /* @__PURE__ */ G("<span class=\"faint\" data-testid=\"no-holder\">waiting for a session</span>"), Is = /* @__PURE__ */ G("<a class=\"mono svelte-t4vask\"> </a>"), Ls = /* @__PURE__ */ G("<span class=\"note svelte-t4vask\"> </span>"), Rs = /* @__PURE__ */ G("<li class=\"svelte-t4vask\"><span> </span> <span class=\"step muted svelte-t4vask\"> </span> <!> <!> <span class=\"faint when svelte-t4vask\"> </span></li>"), zs = /* @__PURE__ */ G("<li class=\"svelte-t4vask\"><span class=\"pill\"> </span><span class=\"step muted svelte-t4vask\"> </span><span class=\"faint\">no report yet</span></li>"), Bs = /* @__PURE__ */ G("<li class=\"svelte-t4vask\"><span class=\"faint\">waiting for a step to start</span></li>"), Vs = /* @__PURE__ */ G("<a class=\"svelte-t4vask\"> </a>"), Hs = /* @__PURE__ */ G("<li class=\"svelte-t4vask\"><span class=\"pill\">comment</span><span class=\"step muted mono svelte-t4vask\"><!></span><span class=\"note svelte-t4vask\"> </span><span class=\"faint when svelte-t4vask\"> </span></li>"), Us = /* @__PURE__ */ G("<div class=\"item card svelte-t4vask\" data-testid=\"progress-card\"><button class=\"head svelte-t4vask\"><span class=\"dot accent\"></span> <span class=\"title svelte-t4vask\"> </span> <span class=\"mono faint id svelte-t4vask\"> </span> <span class=\"faint when svelte-t4vask\"> </span></button> <div class=\"who svelte-t4vask\"><!></div> <ul class=\"svelte-t4vask\"><!> <!></ul></div>"), Ws = /* @__PURE__ */ G("<div class=\"progress svelte-t4vask\" data-testid=\"progress\"></div>"), Gs = /* @__PURE__ */ G("<div class=\"faint\">Loading…</div>"), Ks = /* @__PURE__ */ G("<div class=\"faint none svelte-t4vask\" data-testid=\"progress-empty\">No runs in progress.</div>"), qs = /* @__PURE__ */ G("<header class=\"vh\"><h1>Attention</h1> <span class=\"sub\">What waits on a human, oldest first.</span> <span class=\"actions\"><!></span></header> <div class=\"pad\"><!> <!> <h2 class=\"st\">In progress</h2> <!></div>", 1);
function Js(e, t) {
	Le(t, !0);
	let n = /* @__PURE__ */ F(null), r = /* @__PURE__ */ F(""), i = /* @__PURE__ */ F(Ft([])), a = /* @__PURE__ */ F(Ft({})), o = /* @__PURE__ */ F(!1), s = /* @__PURE__ */ F(""), c = /* @__PURE__ */ F(Ft({
		state: "off",
		error: ""
	}));
	async function l() {
		try {
			I(n, await t.api.get("api/attention"), !0), I(r, "");
		} catch (e) {
			I(r, String(e.message || e), !0);
		}
		try {
			let e = await t.api.get("api/work?state=running&limit=6&detail=1"), n = e.detail ?? {};
			I(a, e.sessions ?? {}, !0), I(i, (e.runs ?? []).slice(0, 6).map((e) => ({
				run: e,
				steps: n[e.id]?.steps ?? [],
				comment: n[e.id]?.comment ?? null
			})), !0);
		} catch {}
		I(o, !0);
	}
	tn(() => {
		l(), t.api.push.status().then((e) => I(c, e, !0));
		let e = 0, n = t.api.subscribe(() => {
			e ||= window.setTimeout(() => {
				e = 0, l();
			}, 1e3);
		}), r = t.api.params();
		if (r.item && !r.session) {
			let [e, ...t] = r.item.split(":");
			u({
				kind: e,
				id: t.join(":")
			});
		}
		return () => {
			n(), clearTimeout(e);
		};
	});
	async function u(e) {
		let n = e.kind + ":" + e.id;
		I(s, n);
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
			I(s, "");
		}
	}
	let d = /* @__PURE__ */ F("");
	async function f(e, r) {
		let i = e.kind + ":" + e.id;
		I(d, i);
		try {
			await Ur(t.api, r) && (I(n, (W(n) ?? []).filter((e) => e.kind + ":" + e.id !== i), !0), l());
		} finally {
			I(d, "");
		}
	}
	let p = {
		unsupported: "This browser has no Web Push (on iOS, add Desk to the home screen first).",
		unavailable: "Desk runs without --push.",
		denied: "Notifications are blocked in the browser settings.",
		off: "Get a notification for each new attention item.",
		on: "Notifications are on for this browser."
	};
	async function m() {
		I(c, W(c).state === "on" ? await t.api.push.unsubscribe() : await t.api.push.subscribe(), !0);
	}
	async function h() {
		I(c, await t.api.push.test(), !0), W(c).error || t.api.toast("Test notification sent");
	}
	function g(e) {
		if (!e || e.startsWith("0001")) return "";
		let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3);
		return t < 60 ? t + "s ago" : t < 3600 ? Math.round(t / 60) + "m ago" : t < 86400 ? Math.round(t / 3600) + "h ago" : Math.round(t / 86400) + "d ago";
	}
	let _ = (e, t) => typeof e?.[t] == "string" ? JSON.parse(e[t] || "{}") : e?.[t] || {};
	function v(e) {
		return e.map((e) => ({
			s: e,
			r: _(e, "meta").report
		})).filter((e) => e.r && (e.r.note || e.r.state)).sort((e, t) => String(t.r.at).localeCompare(String(e.r.at)));
	}
	let y = (e) => e === "failed" || e === "blocked" ? "bad" : e === "waiting" || e === "dropped" ? "warn" : e === "ready" ? "ok" : "accent", b = (e) => !!e && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(e), x = (e) => e?.status === "in_progress" && e.holder ? W(a)[e.holder] ?? { id: e.holder } : null, S = (e) => t.api.openSession(e, { readonly: !0 });
	var C = qs(), w = Ut(C), ee = B(R(w), 4), te = R(ee), T = (e) => {
		var t = Cs(), n = Ut(t), r = (e) => {
			var t = Ss();
			Xn("click", t, h), K(e, t);
		};
		yr(n, (e) => {
			W(c).state === "on" && e(r);
		});
		var i = B(n, 2);
		let a;
		var o = z(i, !0);
		V(() => {
			Fr(i, "title", p[W(c).state]), a = jr(i, 1, "", null, a, { on: W(c).state === "on" }), q(o, W(c).state === "on" ? "Notifications on" : "Notify me");
		}), Xn("click", i, m), K(e, t);
	}, ne = (e) => {
		var t = ws(), n = z(t, !0);
		V(() => {
			Fr(t, "title", p[W(c).state]), q(n, W(c).state === "unavailable" ? "" : "no notifications");
		}), K(e, t);
	};
	yr(te, (e) => {
		W(c).state === "on" || W(c).state === "off" ? e(T) : e(ne, -1);
	}), k(ee), k(w);
	var re = B(w, 2), ie = R(re), ae = (e) => {
		var t = Ts(), n = z(t, !0);
		V(() => q(n, W(c).error)), K(e, t);
	};
	yr(ie, (e) => {
		W(c).error && e(ae);
	});
	var oe = B(ie, 2), se = (e) => {
		var t = Ts(), n = z(t, !0);
		V(() => q(n, W(r))), K(e, t);
	}, ce = (e) => {
		K(e, Es());
	}, le = (e) => {
		K(e, Ds());
	}, ue = (e) => {
		var r = Ns(), i = R(r), a = B(R(i));
		Cr(a, 21, () => W(n), (e) => e.kind + e.id, (e, n) => {
			let r = /* @__PURE__ */ tt(() => W(n).kind + ":" + W(n).id);
			var i = Ms();
			let a;
			var o = R(i), c = z(R(o), !0);
			k(o);
			var l = B(o), p = R(l, !0), m = B(p), h = R(m, !0), _ = B(h), v = (e) => {
				var t = ks(), r = B(Ut(t)), i = (e) => {
					var t = Os(), r = z(t, !0);
					V(() => {
						Fr(t, "href", W(n).url), q(r, W(n).ref);
					}), Xn("click", t, (e) => e.stopPropagation()), K(e, t);
				}, a = /* @__PURE__ */ tt(() => W(n).url && gs(W(n).url)), o = (e) => {
					var t = ar();
					V(() => q(t, W(n).ref)), K(e, t);
				};
				yr(r, (e) => {
					W(a) ? e(i) : e(o, -1);
				}), K(e, t);
			};
			yr(_, (e) => {
				W(n).ref && e(v);
			}), k(m);
			var y = B(m, 2), b = (e) => {
				var t = As(), r = z(t, !0);
				V(() => q(r, W(n).summary)), K(e, t);
			};
			yr(y, (e) => {
				W(n).summary && e(b);
			});
			var x = B(y, 2), S = R(x);
			Cr(S, 17, () => Hr(t.api, W(n).kind, W(n).id), (e) => e.label, (e, t) => {
				var i = js(), a = z(i, !0);
				V(() => {
					i.disabled = W(d) === W(r), q(a, W(t).label);
				}), Xn("click", i, () => f(W(n), W(t))), K(e, i);
			});
			var C = B(S, 2);
			k(x), k(l);
			var w = B(l), ee = z(w, !0), te = z(B(w), !0);
			k(i), V((e) => {
				a = jr(i, 1, "click svelte-t4vask", null, a, { busy: W(s) === W(n).kind + ":" + W(n).id }), Fr(i, "data-item", W(n).kind + ":" + W(n).id), q(c, W(n).kind), q(p, W(n).title || W(n).id), q(h, W(n).id), C.disabled = W(s) === W(r), q(ee, W(n).why), q(te, e);
			}, [() => g(W(n).since)]), Xn("click", i, () => u(W(n))), Xn("click", x, (e) => e.stopPropagation()), Xn("click", C, () => u(W(n))), K(e, i);
		}), k(a), k(i), k(r), K(e, r);
	};
	yr(oe, (e) => {
		W(r) ? e(se) : W(n) ? W(n).length === 0 ? e(le, 2) : e(ue, -1) : e(ce, 1);
	});
	var de = B(oe, 4), fe = (e) => {
		var n = Ws();
		Cr(n, 21, () => W(i), ({ run: e, steps: t, comment: n }) => e.id, (e, n) => {
			let r = () => W(n).run, i = () => W(n).steps, a = () => W(n).comment, o = /* @__PURE__ */ tt(() => v(i())), s = /* @__PURE__ */ tt(() => i().find((e) => e.status === "in_progress")), c = /* @__PURE__ */ tt(() => x(W(s)));
			var l = Us(), u = R(l), d = B(R(u), 2), f = z(d, !0), p = B(d, 2), m = z(p, !0), h = z(B(p, 2), !0);
			k(u);
			var C = B(u, 2), w = R(C), ee = (e) => {
				var t = Ps(), n = Ut(t), r = z(n, !0);
				xs(B(n, 2), {
					get h() {
						return W(c);
					},
					onopen: S
				}), V((e) => q(r, e), [() => _(W(s), "fields").step || W(s).title]), K(e, t);
			}, te = (e) => {
				K(e, Fs());
			};
			yr(w, (e) => {
				W(c) ? e(ee) : e(te, -1);
			}), k(C);
			var T = B(C, 2), ne = R(T);
			Cr(ne, 17, () => W(o).slice(0, 3), ({ s: e, r: t }) => e.id, (e, t) => {
				let n = () => W(t).s, r = () => W(t).r, i = /* @__PURE__ */ tt(() => x(n()));
				var a = Rs(), o = R(a), s = z(o, !0), c = B(o, 2), l = z(c, !0), u = B(c, 2), d = (e) => {
					xs(e, {
						get h() {
							return W(i);
						},
						onopen: S
					});
				}, f = (e) => {
					var t = Is(), n = z(t, !0);
					V(() => {
						Fr(t, "href", "?session=" + r().by), q(n, r().by);
					}), Xn("click", t, (e) => {
						e.preventDefault(), S(r().by);
					}), K(e, t);
				}, p = /* @__PURE__ */ tt(() => b(r().by));
				yr(u, (e) => {
					W(i) ? e(d) : W(p) && e(f, 1);
				});
				var m = B(u, 2), h = (e) => {
					var t = Ls(), n = z(t, !0);
					V(() => q(n, r().note)), K(e, t);
				};
				yr(m, (e) => {
					r().note && e(h);
				});
				var v = z(B(m, 2), !0);
				k(a), V((e, t, n) => {
					jr(o, 1, `pill ${e ?? ""}`, "svelte-t4vask"), q(s, r().state), q(l, t), q(v, n);
				}, [
					() => y(r().state),
					() => _(n(), "fields").step || n().title,
					() => g(r().at)
				]), K(e, a);
			}, (e) => {
				var t = or(), n = Ut(t), r = (e) => {
					var t = zs(), n = R(t), r = z(n, !0), i = z(B(n), !0);
					be(), k(t), V((e) => {
						q(r, W(s).status), q(i, e);
					}, [() => _(W(s), "fields").step || W(s).title]), K(e, t);
				}, i = (e) => {
					K(e, Bs());
				};
				yr(n, (e) => {
					W(s) ? e(r) : e(i, -1);
				}), K(e, t);
			});
			var re = B(ne, 2), ie = (e) => {
				var n = Hs(), r = B(R(n)), i = R(r), o = (e) => {
					var n = Vs(), r = z(n, !0);
					V(() => {
						Fr(n, "href", "?session=" + a().author), q(r, a().author);
					}), Xn("click", n, (e) => {
						e.preventDefault(), t.api.openSession(a().author);
					}), K(e, n);
				}, s = /* @__PURE__ */ tt(() => b(a().author)), c = (e) => {
					var t = ar();
					V(() => q(t, a().author)), K(e, t);
				};
				yr(i, (e) => {
					W(s) ? e(o) : e(c, -1);
				}), k(r);
				var l = B(r), u = z(l, !0), d = z(B(l), !0);
				k(n), V((e) => {
					q(u, a().body), q(d, e);
				}, [() => g(a().created_at)]), K(e, n);
			};
			yr(re, (e) => {
				a() && e(ie);
			}), k(T), k(l), V((e) => {
				Fr(l, "data-run", r().id), q(f, r().title), q(m, r().id), q(h, e);
			}, [() => g(r().updated_at)]), Xn("click", u, () => t.api.openTicket(r().id)), K(e, l);
		}), k(n), K(e, n);
	}, pe = (e) => {
		K(e, Gs());
	}, me = (e) => {
		K(e, Ks());
	};
	yr(de, (e) => {
		W(i).length ? e(fe) : W(o) ? e(me, -1) : e(pe, 1);
	}), k(re), K(e, C), Re();
}
Zn(["click"]);
// plugins/attention/index.svelte.js
var Ys = new URL("./style.css", import.meta.url).href, Xs = class extends HTMLElement {
	#e = /* @__PURE__ */ F(Ft({
		api: null,
		context: {},
		config: {}
	}));
	get props() {
		return W(this.#e);
	}
	set props(e) {
		I(this.#e, e, !0);
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
		if (!document.querySelector(`link[href="${Ys}"]`)) {
			let e = document.createElement("link");
			e.rel = "stylesheet", e.href = Ys, document.head.append(e);
		}
		this.style.display = "contents", this.app = pr(Js, {
			target: this,
			props: this.props
		});
	}
	disconnectedCallback() {
		this.app && _r(this.app), this.app = null;
	}
};
customElements.get("desk-attention-view") || customElements.define("desk-attention-view", Xs);
