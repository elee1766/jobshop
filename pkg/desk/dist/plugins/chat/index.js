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
function y(e, t) {
	if (Array.isArray(e)) return e;
	if (t === void 0 || !(Symbol.iterator in e)) return Array.from(e);
	let n = [];
	for (let r of e) if (n.push(r), n.length === t) break;
	return n;
}
var b = 1024, x = 2048, S = 4096, C = 8192, w = 16384, ee = 32768, te = 1 << 25, T = 65536, E = 1 << 19, ne = 1 << 20, re = 1 << 25, ie = 1 << 21, ae = 1 << 22, oe = 1 << 23, se = Symbol("$state"), ce = Symbol("component"), le = Symbol("legacy props"), ue = Symbol(""), de = Symbol("attributes"), fe = Symbol("class"), pe = Symbol("style"), me = Symbol("text"), D = Symbol("form reset"), he = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), ge = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
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
var O = !1;
function be(e) {
	O = e;
}
var k;
function xe(t) {
	if (t === null) throw ve(), e;
	return k = t;
}
function A() {
	return xe(/* @__PURE__ */ tn(k));
}
function j(t) {
	if (O) {
		if (/* @__PURE__ */ tn(k) !== null) throw ve(), e;
		k = t;
	}
}
function M(e = 1) {
	if (O) {
		for (var t = e, n = k; t--;) n = /* @__PURE__ */ tn(n);
		k = n;
	}
}
function Se(e = !0) {
	for (var t = 0, n = k;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ tn(n);
		e && n.remove(), n = i;
	}
}
function Ce(t) {
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
// node_modules/svelte/src/internal/shared/clone.js
var Re = [];
function ze(e, t = !1, n = !1) {
	return Be(e, /* @__PURE__ */ new Map(), "", Re, null, n);
}
function Be(e, t, n, r, i = null, o = !1) {
	if (typeof e == "object" && e) {
		var s = t.get(e);
		if (s !== void 0) return s;
		if (e instanceof Map) return new Map(e);
		if (e instanceof Set) return new Set(e);
		if (a(e)) {
			var c = Array(e.length);
			t.set(e, c), i !== null && t.set(i, c);
			for (var l = 0; l < e.length; l += 1) {
				var u = e[l];
				l in e && (c[l] = Be(u, t, n, r, null, o));
			}
			return c;
		}
		if (m(e) === f) {
			c = {}, t.set(e, c), i !== null && t.set(i, c);
			for (var d of Object.keys(e)) c[d] = Be(e[d], t, n, r, null, o);
			return c;
		}
		if (e instanceof Date) return e.getTime(), structuredClone(e);
		if (typeof e.toJSON == "function" && !o) return Be(e.toJSON(), t, n, r, e);
	}
	if (e instanceof EventTarget) return e;
	try {
		return structuredClone(e);
	} catch {
		return e;
	}
}
// node_modules/svelte/src/internal/client/context.js
var Ve = null;
function He(e) {
	Ve = e;
}
function Ue(e, t = !1, n) {
	Ve = {
		p: Ve,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: U,
		l: null
	};
}
function We(e) {
	var t = Ve, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) gn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, Ve = t.p, Ge(e);
}
function Ge(e = {}) {
	return l(e, ce, { value: !0 }), e;
}
function Ke() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var qe = [];
function Je() {
	var e = qe;
	qe = [], _(e);
}
function Ye(e) {
	if (qe.length === 0 && !St) {
		var t = qe;
		queueMicrotask(() => {
			t === qe && Je();
		});
	}
	qe.push(e);
}
function Xe() {
	for (; qe.length > 0;) Je();
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Ze = ~(x | S | b);
function N(e, t) {
	e.f = e.f & Ze | t;
}
function Qe(e) {
	e.f & 512 || e.deps === null ? N(e, b) : N(e, S);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function $e(e, t) {
	e.f & 2048 ? t.set(e, x) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, S), N(e, b);
}
// node_modules/svelte/src/internal/client/dom/elements/misc.js
function et(e) {
	O && /* @__PURE__ */ en(e) !== null && rn(e);
}
var tt = !1;
function nt() {
	tt || (tt = !0, document.addEventListener("reset", (e) => {
		Promise.resolve().then(() => {
			if (!e.defaultPrevented) for (let t of e.target.elements) t[D]?.();
		});
	}, { capture: !0 }));
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function rt(e) {
	var t = H, n = U;
	Rn(null), zn(null);
	try {
		return e();
	} finally {
		Rn(t), zn(n);
	}
}
function it(e, t, n, r = n) {
	e.addEventListener(t, () => rt(n));
	let i = e[D];
	e[D] = i ? () => {
		i(), r(!0);
	} : () => r(!0), nt();
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function at(e, t, n, r) {
	let i = Ke() ? lt : P;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = U, c = ot(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				ln(e, s);
			}
			st();
		}
	}
	var d = ct();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ dt(e))).then(u).catch((e) => ln(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), st());
	}) : f();
}
function ot() {
	var e = U, t = H, n = Ve, r = F;
	return function(i = !0) {
		zn(e), Rn(t), He(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function st(e = !0) {
	zn(null), Rn(null), He(null), e && F?.deactivate();
}
function ct() {
	var e = U, t = e.b, n = F, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function lt(e) {
	var n = 2 | x;
	return U !== null && (U.f |= E), {
		ctx: Ve,
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
var ut = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function dt(e, n, r) {
	let i = U;
	i === null && De();
	var a = void 0, o = zt(t), s = !H, c = /* @__PURE__ */ new Set();
	return yn(() => {
		var t = U, n = v();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== he && n.reject(e);
			}).finally(st);
		} catch (e) {
			n.reject(e), st();
		}
		var r = F;
		if (s) {
			if (t.f & 32768) var l = ct();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(ut);
			else for (let e of c.values()) e.reject(ut);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== ut && (r.activate(), t ? (o.f |= oe, Ut(o, t)) : (o.f & 8388608 && (o.f ^= oe), Ut(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), mn(() => {
		for (let e of c) e.reject(ut);
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
function ft(e) {
	let t = /* @__PURE__ */ lt(e);
	return Vn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function P(e) {
	let t = /* @__PURE__ */ lt(e);
	return t.equals = Ee, t;
}
function pt(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) En(t[n]);
	}
}
function mt(e) {
	var n, r = U, i = e.parent;
	if (!Fn && i !== null && e.v !== t && i.f & 24576) return _e(), e.v;
	zn(i);
	try {
		pt(e), n = $n(e);
	} finally {
		zn(r);
	}
	return n;
}
function ht(e) {
	var t = mt(e);
	!e.equals(t) && (e.wv = Xn(), (!F?.is_fork || e.deps === null) && (F === null ? e.v = t : (F.capture(e, t, !0), yt?.capture(e, t, !0)), e.deps === null)) ? N(e, b) : Fn || (bt === null ? Qe(e) : (pn() || F?.is_fork) && bt.set(e, t));
}
function gt(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && rt(() => {
		t.ac.abort(he), t.ac = null;
	}), t.fn !== null && (t.teardown = g), nr(t, 0), wn(t));
}
function _t(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && rr(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var vt = null, F = null, yt = null, bt = null, xt = null, St = !1, Ct = !1, wt = null, Tt = null, Et = 0, Dt = 1, Ot = class e {
	id = Dt++;
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
		vt === null ? vt = this : (vt.next = this, this.prev = vt), vt = this;
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
			for (var r of n.d) N(r, x), t(r);
			for (r of n.m) N(r, S), t(r);
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
					t.f ^= b;
				}
			}
			n || e.push(t);
		}
		return this.#a = [], e;
	}
	#m() {
		this.started = !0;
		for (let [e, t] of this.#s) N(e, t), this.schedule(e);
		this.apply();
		for (var t = wt = [], n = [], r = Tt = []; this.#a.length > 0;) {
			Et++ > 1e3 && (this.#b(), At());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Ft(e), this.#d() || this.discard(), t;
			}
		}
		if (F = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (wt = null, Tt = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Pt(e, t);
			r.length > 0 && F.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), yt = this, Mt(n), Mt(t), yt = null, this.#i?.resolve();
			var o = F;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Lt.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= b;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= b : i & 4 ? t.push(r) : Zn(r) && (i & 16 && this.add_dirty_reaction(r, S), rr(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), N(i, x), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), F = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) $e(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), bt?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		F = this;
	}
	deactivate() {
		F = null, bt = null;
	}
	flush() {
		try {
			Ct = !0, F = this, this.#m();
		} finally {
			Et = 0, xt = null, wt = null, Tt = null, Ct = !1, F = null, bt = null, Lt.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(ut);
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
		this.#u || (this.#u = !0, Ye(() => {
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
		if (F === null) {
			let t = F = new e();
			!Ct && !St && Ye(() => {
				t.started || t.flush();
			});
		}
		return F;
	}
	apply() {
		bt = null;
	}
	schedule(e) {
		xt = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? vt = e : t.prev = e, this.linked = !1;
		}
	}
};
function kt(e) {
	var t = St, n = yt;
	yt = null, St = !0;
	try {
		var r;
		for (e && (kt(), r = e());;) {
			if (Xe(), F === null) return r;
			F.flush();
		}
	} finally {
		St = t, yt = n;
	}
}
function At() {
	try {
		Me();
	} catch (e) {
		ln(e, xt);
	}
}
var jt = null;
function Mt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Zn(r) && (jt = /* @__PURE__ */ new Set(), rr(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && On(r), jt?.size > 0)) {
				Lt.clear();
				for (let e of jt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) jt.has(n) && (jt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || rr(n);
					}
				}
				jt.clear();
			}
		}
		jt = null;
	}
}
function Nt(e) {
	F.schedule(e);
}
function Pt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), N(e, b);
		for (var n = e.first; n !== null;) Pt(n, t), n = n.next;
	}
}
function Ft(e) {
	N(e, b);
	for (var t = e.first; t !== null;) Ft(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var It = /* @__PURE__ */ new Set(), Lt = /* @__PURE__ */ new Map(), Rt = !1;
function zt(e, t) {
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
function I(e, t) {
	let n = zt(e, t);
	return Vn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Bt(e, t = !1, n = !0) {
	let r = zt(e);
	return t || (r.equals = Ee), r;
}
function L(e, t, n = !1) {
	return H !== null && (!Ln || H.f & 131072) && Ke() && H.f & 4325394 && (Bn === null || !Bn.has(e)) && Ie(), Ut(e, n ? qt(t) : t, Tt);
}
var Vt = null, Ht = 0;
function Ut(e, t, n = null) {
	if (!e.equals(t)) {
		Fn ? Lt.set(e, t) : Lt.has(e) || Lt.set(e, e.v);
		var r = Ot.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && mt(t), bt === null && Qe(t);
		}
		e.wv = Xn(), Vt = null, Ht = 0, Kt(e, x, n), Vt = null, Ke() && U !== null && U.f & 1024 && !(U.f & 96) && (Wn === null ? Gn([e]) : Wn.push(e)), !r.is_fork && It.size > 0 && !Rt && Wt();
	}
	return t;
}
function Wt() {
	Rt = !1;
	for (let e of It) {
		e.f & 1024 && N(e, S);
		let t;
		try {
			t = Zn(e);
		} catch {
			t = !0;
		}
		t && rr(e);
	}
	It.clear();
}
function Gt(e) {
	L(e, e.v + 1);
}
function Kt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ke(), a = r.length;
		if (Ht += a, Ht > 1e5 && Vt === null && (Vt = /* @__PURE__ */ new Set()), Vt !== null) {
			if (Vt.has(e)) return;
			Vt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== U) {
				var l = (c & x) === 0;
				if (l && N(s, t), c & 131072) It.add(s);
				else if (c & 2) {
					var u = s;
					bt?.delete(u), Kt(u, S, n);
				} else if (l) {
					var d = s;
					c & 16 && jt !== null && jt.add(d), n === null ? Nt(d) : n.push(d);
				}
			}
		}
	}
}
function qt(e) {
	if (typeof e != "object" || !e || se in e || ce in e) return e;
	let n = m(e);
	if (n !== f && n !== p) return e;
	var r = /* @__PURE__ */ new Map(), i = a(e), o = /* @__PURE__ */ I(0), s = null, c = Jn, l = (e) => {
		if (Jn === c) return e();
		var t = H, n = Jn;
		Rn(null), Yn(c);
		var r = e();
		return Rn(t), Yn(n), r;
	};
	return i && r.set("length", /* @__PURE__ */ I(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Pe();
			var i = r.get(t);
			return i === void 0 ? l(() => {
				var e = /* @__PURE__ */ I(n.value, s);
				return r.set(t, e), e;
			}) : L(i, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var i = r.get(n);
			if (i === void 0) {
				if (n in e) {
					let e = l(() => /* @__PURE__ */ I(t, s));
					r.set(n, e), Gt(o);
				}
			} else L(i, t), Gt(o);
			return !0;
		},
		get(n, i, a) {
			if (i === se) return e;
			var o = r.get(i), c = i in n;
			if (o === void 0 && (!c || u(n, i)?.writable) && (o = l(() => /* @__PURE__ */ I(qt(c ? n[i] : t), s)), r.set(i, o)), o !== void 0) {
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
			if (n === se) return !0;
			var i = r.get(n), a = i !== void 0 && i.v !== t || Reflect.has(e, n);
			return (i !== void 0 || (U !== null || H !== null) && (!a || u(e, n)?.writable)) && (i === void 0 && (i = l(() => /* @__PURE__ */ I(a ? qt(e[n]) : t, s)), r.set(n, i)), W(i) === t) ? !1 : a;
		},
		set(e, n, a, c) {
			var d = r.get(n), f = n in e;
			if (i && n === "length") for (var p = a; p < d.v; p += 1) {
				var m = r.get(p + "");
				m === void 0 ? p in e && (m = l(() => /* @__PURE__ */ I(t, s)), r.set(p + "", m)) : L(m, t);
			}
			if (d === void 0) (!f || u(e, n)?.writable) && (d = l(() => /* @__PURE__ */ I(void 0, s)), L(d, qt(a)), r.set(n, d));
			else {
				f = d.v !== t;
				var h = l(() => qt(a));
				L(d, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, a), !f) {
				if (i && typeof n == "string") {
					var _ = r.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && L(_, v + 1);
				}
				Gt(o);
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
var Jt, Yt, Xt, Zt;
function Qt() {
	if (Jt === void 0) {
		Jt = window, Yt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Xt = u(t, "firstChild").get, Zt = u(t, "nextSibling").get, h(e) && (e[fe] = void 0, e[de] = null, e[pe] = void 0, e.__e = void 0), h(n) && (n[me] = void 0);
	}
}
function $t(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function en(e) {
	return Xt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function tn(e) {
	return Zt.call(e);
}
function R(e, t) {
	if (!O) return /* @__PURE__ */ en(e);
	var n = /* @__PURE__ */ en(k);
	if (n === null) n = k.appendChild($t());
	else if (t && n.nodeType !== 3) {
		var r = $t();
		return n?.before(r), xe(r), r;
	}
	return t && sn(n), xe(n), n;
}
function nn(e, t = !1) {
	if (!O) {
		var n = /* @__PURE__ */ en(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ tn(n) : n;
	}
	if (t) {
		if (k?.nodeType !== 3) {
			var r = $t();
			return k?.before(r), xe(r), r;
		}
		sn(k);
	}
	return k;
}
function z(e, t = !1) {
	if (!O) return /* @__PURE__ */ en(e);
	var n = R(e, t);
	return j(e), n;
}
function B(e, t = 1, n = !1) {
	let r = O ? k : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ tn(r);
	if (!O) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = $t();
			return r === null ? i?.after(a) : r.before(a), xe(a), a;
		}
		sn(r);
	}
	return xe(r), r;
}
function rn(e) {
	e.textContent = "";
}
function an() {
	return !1;
}
function on(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function sn(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function cn(e) {
	var t = U;
	if (t === null) return H.f |= oe, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	ln(e, t);
}
function ln(t, n) {
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
function un(e) {
	U === null && (H === null && je(e), Ae()), Fn && ke(e);
}
function dn(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function fn(e, t) {
	var n = U;
	n !== null && n.f & 8192 && (e |= C);
	var r = {
		ctx: Ve,
		deps: null,
		nodes: null,
		f: e | x | 512,
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
	F?.register_created_effect(r);
	var i = r;
	if (e & 4) wt === null ? Ot.ensure().schedule(r) : wt.push(r);
	else if (t !== null) {
		try {
			rr(r);
		} catch (e) {
			throw En(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= T));
	}
	if (i !== null && (i.parent = n, n !== null && dn(i, n), H !== null && H.f & 2 && !(e & 64))) {
		var a = H;
		(a.effects ??= []).push(i);
	}
	return r;
}
function pn() {
	return H !== null && !Ln;
}
function mn(e) {
	let t = fn(8, null);
	return N(t, b), t.teardown = e, t;
}
function hn(e) {
	un("$effect");
	var t = U.f;
	if (!H && t & 32 && Ve !== null && !Ve.i) {
		var n = Ve;
		(n.e ??= []).push(e);
	} else return gn(e);
}
function gn(e) {
	return fn(4 | ne, e);
}
function _n(e) {
	Ot.ensure();
	let t = fn(64 | E, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? kn(t, () => {
			En(t), n(void 0);
		}) : (En(t), n(void 0));
	});
}
function vn(e) {
	return fn(4, e);
}
function yn(e) {
	return fn(ae | E, e);
}
function bn(e, t = 0) {
	return fn(8 | t, e);
}
function V(e, t = [], n = [], r = []) {
	at(r, t, n, (t) => {
		fn(8, () => {
			e(...t.map(W));
		});
	});
}
function xn(e, t = 0) {
	return fn(16 | t, e);
}
function Sn(e) {
	return fn(32 | E, e);
}
function Cn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = Fn, r = H;
		In(!0), Rn(null);
		try {
			t.call(null);
		} catch (t) {
			ln(t, e.parent);
		} finally {
			In(n), Rn(r);
		}
	}
}
function wn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && rt(() => {
			e.abort(he);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : En(n, t), n = r;
	}
}
function Tn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || En(t), t = n;
	}
}
function En(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (Dn(e.nodes.start, e.nodes.end), n = !0), e.f |= te, wn(e, t && !n), nr(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	Cn(e), e.f ^= te, e.f |= w;
	var i = e.parent;
	i !== null && i.first !== null && On(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function Dn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ tn(e);
		e.remove(), e = n;
	}
}
function On(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function kn(e, t, n = !0) {
	var r = [];
	e.f |= 256, An(e, r, !0);
	var i = () => {
		n && En(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function An(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= C;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				An(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function jn(e) {
	e.f &= -257, Mn(e, !0);
}
function Mn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= C, e.f & 1024 || (N(e, x), Ot.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			Mn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function Nn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ tn(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var Pn = null, Fn = !1;
function In(e) {
	Fn = e;
}
var H = null, Ln = !1;
function Rn(e) {
	H = e;
}
var U = null;
function zn(e) {
	U = e;
}
var Bn = null;
function Vn(e) {
	H !== null && (H.f & 2097152 || H.f & 2) && (Bn ??= /* @__PURE__ */ new Set()).add(e);
}
var Hn = null, Un = 0, Wn = null;
function Gn(e) {
	Wn = e;
}
var Kn = 1, qn = 0, Jn = qn;
function Yn(e) {
	Jn = e;
}
function Xn() {
	return ++Kn;
}
function Zn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Zn(a) && ht(a), a.wv > e.wv) return !0;
		}
		t & 512 && bt === null && N(e, b);
	}
	return !1;
}
function Qn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(Bn !== null && Bn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Qn(a, t, !1) : t === a && (n ? N(a, x) : a.f & 1024 && N(a, S), Nt(a));
	}
}
function $n(e) {
	var t = Hn, n = Un, r = Wn, i = H, a = Bn, o = Ve, s = Ln, c = Jn, l = e.f;
	Hn = null, Un = 0, Wn = null, H = l & 96 ? null : e, Bn = null, He(e.ctx), Ln = !1, Jn = ++qn, e.ac !== null && (rt(() => {
		e.ac.abort(he);
	}), e.ac = null);
	try {
		e.f |= ie;
		var u = e.fn, d = u();
		e.f |= ee;
		var f = er(e);
		if (Ke() && Wn !== null && !Ln && f !== null && !(e.f & 6146)) for (var p = 0; p < Wn.length; p++) Qn(Wn[p], e);
		if (i !== null && i !== e) {
			if (qn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = qn;
			if (t !== null) for (let e of t) e.rv = qn;
			Wn !== null && (r === null ? r = Wn : r.push(...Wn));
		}
		return e.f & 8388608 && (e.f ^= oe), d;
	} catch (t) {
		return er(e), cn(t);
	} finally {
		e.f ^= ie, Hn = t, Un = n, Wn = r, H = i, Bn = a, He(o), Ln = s, Jn = c;
	}
}
function er(e) {
	var t = e.deps, n = F?.is_fork;
	if (Hn !== null) {
		var r;
		if (n || nr(e, Un), t !== null && Un > 0) for (t.length = Un + Hn.length, r = 0; r < Hn.length; r++) t[Un + r] = Hn[r];
		else e.deps = t = Hn;
		if (pn() && e.f & 512) for (r = Un; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && Un < t.length && (nr(e, Un), t.length = Un);
	return t;
}
function tr(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var i = o.call(r, e);
		if (i !== -1) {
			var a = r.length - 1;
			a === 0 ? r = n.reactions = null : (r[i] = r[a], r.pop());
		}
	}
	if (r === null && n.f & 2 && (Hn === null || !s.call(Hn, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Qe(c), c.ac !== null && rt(() => {
			c.ac.abort(he), c.ac = null, N(c, x);
		}), gt(c), nr(c, 0);
	}
}
function nr(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) tr(e, n[r]);
}
function rr(e) {
	var t = e.f;
	if (!(t & 16384)) {
		N(e, b);
		var n = U;
		U = e;
		try {
			t & 16777232 ? Tn(e) : wn(e), Cn(e);
			var r = $n(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Kn;
		} finally {
			U = n;
		}
	}
}
async function ir() {
	await Promise.resolve(), kt();
}
function W(e) {
	var t = !!(e.f & 2);
	if (Pn?.add(e), H !== null && !Ln && !(U !== null && U.f & 16384) && (Bn === null || !Bn.has(e))) {
		var n = H.deps;
		if (H.f & 2097152) e.rv < qn && (e.rv = qn, Hn === null && n !== null && n[Un] === e ? Un++ : Hn === null ? Hn = [e] : Hn.push(e));
		else {
			H.deps ??= [], s.call(H.deps, e) || H.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [H] : s.call(r, H) || r.push(H);
		}
	}
	if (Fn && Lt.has(e)) return Lt.get(e);
	if (t) {
		var i = e;
		if (Fn) {
			var a = i.v;
			return (!(i.f & 1024) && i.reactions !== null || or(i)) && (a = mt(i)), Lt.set(i, a), a;
		}
		var o = !(i.f & 512) && !Ln && H !== null && !!(H.f & 512), c = (i.f & ee) === 0;
		Zn(i) && (o && (i.f |= 512), ht(i)), o && !c && (_t(i), ar(i));
	}
	if (bt?.has(e)) return bt.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function ar(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : s.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (_t(n), ar(n));
	}
}
function or(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Lt.has(t) || t.f & 2 && or(t)) return !0;
	return !1;
}
function sr(e) {
	var t = Ln;
	try {
		return Ln = !0, e();
	} finally {
		Ln = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var cr = Symbol("events"), lr = /* @__PURE__ */ new Set(), ur = /* @__PURE__ */ new Set();
function dr(e, t, n, r = {}) {
	function i(e) {
		if (r.capture || _r.call(t, e), !e.cancelBubble) return rt(() => n?.call(this, e));
	}
	return e.startsWith("pointer") || e.startsWith("touch") || e === "wheel" ? (i.__removed = !1, Ye(() => {
		i.__removed || t.addEventListener(e, i, r);
	})) : t.addEventListener(e, i, r), i;
}
function fr(e, t, n, r, i) {
	var a = {
		capture: r,
		passive: i
	}, o = dr(e, t, n, a);
	(t === document.body || t === window || t === document || t instanceof HTMLMediaElement) && mn(() => {
		o.__removed = !0, t.removeEventListener(e, o, a);
	});
}
function pr(e, t, n) {
	(t[cr] ??= {})[e] = n;
}
function mr(e) {
	for (var t = 0; t < e.length; t++) lr.add(e[t]);
	for (var n of ur) n(e);
}
var hr = null, gr = !1;
function _r(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	hr = e, gr || (gr = !0, setTimeout(() => {
		gr = !1, hr = null;
	}));
	var o = 0, s = hr === e && e[cr];
	if (s) {
		var c = i.indexOf(s);
		if (c !== -1 && (t === document || t === window)) {
			e[cr] = t;
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
		Rn(null), zn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[cr]?.[r];
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
			e[cr] = t, delete e.currentTarget, Rn(d), zn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var vr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function yr(e) {
	return vr?.createHTML(e) ?? e;
}
function br(e) {
	var t = on("template");
	return t.innerHTML = yr(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function xr(e, t) {
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
		if (O) return xr(k, null), k;
		i === void 0 && (i = br(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ en(i)));
		var t = r || Yt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ en(t), s = t.lastChild;
			xr(o, s);
		} else xr(t, t);
		return t;
	};
}
function Sr(e = "") {
	if (!O) {
		var t = $t(e + "");
		return xr(t, t), t;
	}
	var n = k;
	return n.nodeType === 3 ? sn(n) : (n.before(n = $t()), xe(n)), xr(n, n), n;
}
function Cr() {
	if (O) return xr(k, null), k;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = $t();
	return e.append(t, n), xr(t, n), e;
}
function K(e, t) {
	if (O) {
		var n = U;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = k), A();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var wr = ["touchstart", "touchmove"];
function Tr(e) {
	return wr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function Er(e) {
	let t = 0, n = zt(0), r;
	return () => {
		pn() && (W(n), bn(() => (t === 0 && (r = sr(() => e(() => Gt(n)))), t += 1, () => {
			Ye(() => {
				--t, t === 0 && (r?.(), r = void 0, Gt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var Dr = T | E;
function Or(e, t, n, r) {
	new kr(e, t, n, r);
}
var kr = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = O ? k : null;
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
	#m = Er(() => (this.#p = zt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = U;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = U.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = xn(() => {
			if (O) {
				let e = Ce(this.#t);
				A();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, Dr), O && (this.#e = k);
	}
	#h() {
		try {
			this.#a = Sn(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		Ye(r), t && (this.#s = Sn(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? ye() : (t = !0, n && Le(), this.#s !== null && kn(this.#s, () => {
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
					ln(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (w | te)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = Sn(() => e(this.#e)), Ye(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = $t(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return Sn(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							ln(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(F)) : this.#u === 0 && (this.#e.before(e), this.#c = null, kn(this.#o, () => {
					this.#o = null;
				}), this.#x(F));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = Sn(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				Nn(this.#a, e);
				let t = this.#n.pending;
				this.#o = Sn(() => t(this.#e));
			} else this.#x(F);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		$e(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = U, n = H, r = Ve;
		zn(this.#i), Rn(this.#i), He(this.#i.ctx);
		try {
			return Ot.ensure(), e();
		} finally {
			zn(t), Rn(n), He(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && kn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Ye(() => {
			this.#d = !1, this.#p && Ut(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), W(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		F?.is_fork ? (this.#a && F.skip_effect(this.#a), this.#o && F.skip_effect(this.#o), this.#s && F.skip_effect(this.#s), F.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (En(this.#a), null), this.#o &&= (En(this.#o), null), this.#s &&= (En(this.#s), null), O && (xe(this.#t), M(), xe(Se()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return Sn(() => {
						var r = U;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return ln(e, this.#i.parent), null;
				}
			}));
		};
		Ye(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					ln(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => ln(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function q(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[me] ??= e.nodeValue) && (e[me] = n, e.nodeValue = `${n}`);
}
function Ar(e, t) {
	return Mr(e, t);
}
var jr = /* @__PURE__ */ new Map();
function Mr(t, { target: n, anchor: r, props: i = {}, events: a, context: o, intro: s = !0, transformError: l }) {
	Qt();
	var u = void 0, d = _n(() => {
		var s = r ?? n.appendChild($t());
		Or(s, { pending: () => {} }, (n) => {
			Ue({});
			var r = Ve;
			if (o && (r.c = o), a && (i.$$events = a), O && xr(n, null), u = t(n, i) || Ge(), O && (U.nodes.end = k, k === null || k.nodeType !== 8 || k.data !== "]")) throw ve(), e;
			We();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = Tr(r);
					for (let e of [n, document]) {
						var a = jr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), jr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, _r, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(c(lr)), ur.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = jr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, _r), t.delete(e), t.size === 0 && jr.delete(r)) : t.set(e, i);
			}
			ur.delete(f), s !== r && s.parentNode?.removeChild(s);
		};
	});
	return Nr.set(u, d), u;
}
var Nr = /* @__PURE__ */ new WeakMap();
function Pr(e, t) {
	let n = Nr.get(e);
	return n ? (Nr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var Fr = class {
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
			if (n) jn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (jn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (En(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						Nn(r, t), t.append($t()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else En(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), kn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (En(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = F, r = an();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = $t();
				i.append(a), this.#n.set(e, {
					effect: Sn(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, Sn(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else O && (this.anchor = k), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function J(e, t, n = !1) {
	var r;
	O && (r = k, A());
	var i = new Fr(e), a = n ? T : 0;
	function o(e, t) {
		if (O) {
			var n = Ce(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Se();
				xe(a), i.anchor = a, be(!1), i.ensure(e, t), be(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	xn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function Ir(e, t) {
	return t;
}
function Lr(e, t, n) {
	for (var r = [], i = t.length, a, o = t.length, s = 0; s < i; s++) {
		let n = t[s];
		kn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					Rr(e, c(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --o;
		}, !1);
	}
	if (o === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			rn(d), d.append(u), e.items.clear();
		}
		Rr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function Rr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= re, Nn(a, document.createDocumentFragment())) : En(t[i], n);
	}
}
var zr;
function Br(e, t, n, r, i, o = null) {
	var s = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		s = O ? xe(/* @__PURE__ */ en(u)) : u.appendChild($t());
	}
	O && A();
	var d = null, f = /* @__PURE__ */ P(() => {
		var e = n();
		return a(e) ? e : e == null ? [] : c(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = W(f);
			_.fallback = d, Hr(_, n, s, t, r), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= re, Wr(d, null, s)) : jn(d) : kn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: xn(() => {
			var e = W(f), a = e.length;
			let c = !1;
			O && Ce(s) === "[!" != (a === 0) && (s = Se(), xe(s), be(!1), c = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = F, v = an(), y = 0; y < a; y += 1) {
				O && k.nodeType === 8 && k.data === "]" && (s = k, c = !0, be(!1));
				var b = e[y], x = r(b, y), S = m ? null : l.get(x);
				S ? (S.v && Ut(S.v, b), S.i && Ut(S.i, y), v && _.unskip_effect(S.e)) : (S = Ur(l, m ? s : zr ??= $t(), b, x, y, i, t, n), m || (S.e.f |= re), l.set(x, S)), u.add(x);
			}
			if (a === 0 && o && !d && (m ? d = Sn(() => o(s)) : (d = Sn(() => o(zr ??= $t())), d.f |= re)), a > u.size && Oe("", "", ""), O && a > 0 && xe(Se()), !m) {
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
	m = !1, O && (s = k);
}
function Vr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function Hr(e, t, n, r, i) {
	var a = !!(r & 8), o = t.length, s = e.items, l = Vr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < o; v += 1) h = t[v], g = i(h, v), _ = s.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < o; v += 1) {
		if (h = t[v], g = i(h, v), _ = s.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (jn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= re, _ === l) Wr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), Gr(e, d, _), Gr(e, _, y), Wr(_, y, n), d = _, p = [], m = [], l = Vr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) Wr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					Gr(e, S.prev, C.next), Gr(e, d, S), Gr(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), Wr(_, l, n), Gr(e, _.prev, _.next), Gr(e, _, d === null ? e.effect.first : d.next), Gr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = Vr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = Vr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (Rr(e, c(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var w = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || w.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && w.push(l), l = Vr(l.next);
		var ee = w.length;
		if (ee > 0) {
			var te = r & 4 && o === 0 ? n : null;
			if (a) {
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.measure();
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.fix();
			}
			Lr(e, w, te);
		}
	}
	a && Ye(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Ur(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? zt(n) : /* @__PURE__ */ Bt(n, !1, !1) : null, l = o & 2 ? zt(i) : null;
	return {
		v: c,
		i: l,
		e: Sn(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Wr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ tn(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function Gr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
function Kr(t, n, a = !1, o = !1, s = !1, c = !1) {
	var l = t, u = "";
	if (a) {
		var d = t;
		O && (l = xe(/* @__PURE__ */ en(d)));
	}
	V(() => {
		var t = U;
		if (u === (u = n() ?? "")) O && A();
		else if (a && !O) t.nodes = null, d.innerHTML = u, u !== "" && xr(/* @__PURE__ */ en(d), d.lastChild);
		else if (t.nodes !== null && (Dn(t.nodes.start, t.nodes.end), t.nodes = null), u !== "") {
			if (O) {
				for (var c = k.data, f = A(), p = f; f !== null && (f.nodeType !== 8 || f.data !== "");) p = f, f = /* @__PURE__ */ tn(f);
				if (f === null) throw ve(), e;
				xr(k, p), l = xe(f);
			} else {
				var m = on(o ? "svg" : s ? "math" : "template", o ? r : s ? i : void 0);
				m.innerHTML = u;
				var h = o || s ? m : m.content;
				if (xr(/* @__PURE__ */ en(h), h.lastChild), o || s) for (; /* @__PURE__ */ en(h);) l.before(/* @__PURE__ */ en(h));
				else l.before(h);
			}
		}
	});
}
// node_modules/svelte/src/internal/shared/attributes.js
var qr = [..." 	\n\r\f\xA0\v﻿"];
function Jr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || qr.includes(r[o - 1])) && (s === r.length || qr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Yr(e, t, n, r, i, a) {
	var o = e[fe];
	if (O || o !== n || o === void 0) {
		var s = Jr(n, r, a);
		O && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[fe] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Xr = Symbol("is custom element"), Zr = Symbol("is html"), Qr = ge ? "link" : "LINK";
function $r(e) {
	if (O) {
		var t = !1, n = () => {
			if (!t) {
				if (t = !0, e.hasAttribute("value")) {
					var n = e.value;
					ei(e, "value", null), e.value = n;
				}
				if (e.hasAttribute("checked")) {
					var r = e.checked;
					ei(e, "checked", null), e.checked = r;
				}
			}
		};
		e[D] = n, Ye(n), nt();
	}
}
function ei(e, t, n, r) {
	var i = ti(e);
	if (!(O && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Qr))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[ue] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && ri(e).has(t) ? e[t] = n : (!O || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function ti(e) {
	return e[de] ??= {
		[Xr]: e.nodeName.includes("-"),
		[Zr]: e.namespaceURI === n
	};
}
var ni = /* @__PURE__ */ new Map();
function ri(e) {
	var t = e.getAttribute("is") || e.nodeName, n = ni.get(t);
	if (n) return n;
	ni.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = d(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = m(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/input.js
function ii(e, t, n = t) {
	var r = /* @__PURE__ */ new WeakSet();
	it(e, "input", async (i) => {
		var a = i ? e.defaultValue : e.value;
		if (a = oi(e) ? si(a) : a, n(a), F !== null && r.add(F), await ir(), a !== (a = t())) {
			var o = e.selectionStart, s = e.selectionEnd, c = e.value.length;
			if (e.value = a ?? "", s !== null) {
				var l = e.value.length;
				o === s && s === c && l > c ? (e.selectionStart = l, e.selectionEnd = l) : (e.selectionStart = o, e.selectionEnd = Math.min(s, l));
			}
		}
	}), (O && e.defaultValue !== e.value || sr(t) == null && e.value) && (n(oi(e) ? si(e.value) : e.value), F !== null && r.add(F)), bn(() => {
		var n = t();
		if (e === document.activeElement) {
			var i = F;
			if (r.has(i)) return;
		}
		oi(e) && n === si(e.value) || (e.type !== "date" || n || e.value) && n !== e.value && (e.value = n ?? "");
	});
}
function ai(e, t, n = t) {
	it(e, "change", (t) => {
		n(t ? e.defaultChecked : e.checked);
	}), (O && e.defaultChecked !== e.checked || sr(t) == null) && n(e.checked), bn(() => {
		e.checked = !!t();
	});
}
function oi(e) {
	var t = e.type;
	return t === "number" || t === "range";
}
function si(e) {
	return e === "" ? null : +e;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/this.js
function ci(e, t) {
	return e === t || e?.[se] === t;
}
function li(e = Ge(), t, n, r) {
	var i = Ve.r, a = U;
	return vn(() => {
		var o, s;
		return bn(() => {
			o = s, s = r?.() || [], sr(() => {
				ci(n(...s), e) || (t(e, ...s), o && ci(n(...o), e) && t(null, ...o));
			});
		}), () => {
			let r = a;
			for (; r !== i && r.parent !== null && r.parent.f & 33554432;) r = r.parent;
			let o = () => {
				s && ci(n(...s), e) && t(null, ...s);
			}, c = r.teardown;
			r.teardown = () => {
				o(), c?.();
			};
		};
	}), e;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var ui = !1;
function di(e) {
	var t = ui;
	try {
		return ui = !1, [e(), ui];
	} finally {
		ui = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
function fi(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, c = !0, l = void 0, d = () => o && i ? (l ??= /* @__PURE__ */ lt(r), W(l)) : (c && (c = !1, s = o ? sr(r) : r), s);
	let f;
	if (a) {
		var p = se in e || le in e;
		f = u(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = di(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Ne(t), f(m)));
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
	var v = !1, y = (n & 1 ? lt : P)(() => (v = !1, g()));
	a && W(y);
	var b = U;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? W(y) : i && a ? qt(e) : e;
			return L(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return Fn && v || b.f & 16384 ? y.v : W(y);
	});
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// plugins/chat/transcript.svelte.js
var pi = "Messages for you:";
function mi(e) {
	switch (e) {
		case "pending": return "input";
		case "in_progress": return "running";
		case "completed": return "done";
		case "failed": return "failed";
		default: return e || "input";
	}
}
function hi(e) {
	if (!e.startsWith(pi)) return {
		source: "controller",
		text: e
	};
	let t = /\n\nFrom ([^\s(]+) \(/.exec(e), n = t ? t[1] : "", r = e.slice(17).replace(/^\s*\n\nFrom [^\n]*\n/, "").replace(/\n\nFrom [^\n]*\n/g, "\n\n");
	return {
		source: n.startsWith("desk:") ? "desk" : n ? "operator" : "message",
		text: r.trim(),
		who: n
	};
}
var gi = class {
	#e = /* @__PURE__ */ I(qt([]));
	get blocks() {
		return W(this.#e);
	}
	set blocks(e) {
		L(this.#e, e, !0);
	}
	#t = /* @__PURE__ */ I(!1);
	get busy() {
		return W(this.#t);
	}
	set busy(e) {
		L(this.#t, e, !0);
	}
	#n = /* @__PURE__ */ I(null);
	get usage() {
		return W(this.#n);
	}
	set usage(e) {
		L(this.#n, e, !0);
	}
	last(e) {
		let t = this.blocks[this.blocks.length - 1];
		return t && t.kind === e ? t : void 0;
	}
	tool(e) {
		for (let t = this.blocks.length - 1; t >= 0; t--) {
			let n = this.blocks[t];
			if (n.kind === "tool" && n.id === e) return n;
		}
	}
	settle() {
		for (let e of this.blocks) (e.kind === "assistant" || e.kind === "reasoning") && e.live && (e.live = !1);
	}
	apply(e) {
		let t = e.seq || 0;
		switch (e.type) {
			case "turn_start": {
				this.settle(), this.busy = !0;
				let n = hi(e.text || "");
				e.context && this.blocks.push({
					kind: "system",
					text: "Context given to the session:\n" + e.context,
					seq: t
				}), n.text && this.blocks.push({
					kind: "user",
					text: n.text,
					source: n.source,
					context: e.context || "",
					seq: t
				});
				break;
			}
			case "text": {
				this.busy = !0;
				let n = this.last("assistant");
				n?.live ? n.text += e.text : this.blocks.push({
					kind: "assistant",
					text: e.text || "",
					live: !0,
					seq: t
				});
				break;
			}
			case "reasoning": {
				this.busy = !0;
				let n = this.last("reasoning");
				n?.live ? n.text += e.text : this.blocks.push({
					kind: "reasoning",
					text: e.text || "",
					live: !0,
					seq: t
				});
				break;
			}
			case "tool": {
				let n = e.tool;
				if (!n) break;
				this.busy = !0;
				let r = n.id || n.title, i = this.tool(r);
				i ? (n.title && (i.name = n.title), i.state = mi(n.status), n.input && (i.input = n.input), n.output && (i.output = n.output)) : (this.settle(), this.blocks.push({
					kind: "tool",
					id: r,
					name: n.title || r,
					input: n.input || "",
					output: n.output || "",
					state: mi(n.status),
					seq: t
				}));
				break;
			}
			case "turn_end":
				this.settle(), this.busy = !1, e.error ? this.blocks.push({
					kind: "error",
					text: "Turn failed: " + e.error,
					seq: t
				}) : e.stop_reason === "cancelled" ? this.blocks.push({
					kind: "system",
					text: "Stopped",
					seq: t
				}) : e.stop_reason && e.stop_reason !== "end_turn" && this.blocks.push({
					kind: "system",
					text: "Turn ended: " + e.stop_reason,
					seq: t
				});
				break;
			case "usage":
				e.usage && (this.usage = e.usage);
				break;
			case "state":
				e.status === "closed" && (this.busy = !1, this.blocks.push({
					kind: "system",
					text: "Session closed" + (e.resolution ? " · " + e.resolution : ""),
					seq: t
				}));
				break;
			default: return !1;
		}
		return !0;
	}
}, _i = class {
	#e = /* @__PURE__ */ I("connecting");
	get status() {
		return W(this.#e);
	}
	set status(e) {
		L(this.#e, e, !0);
	}
	#t = /* @__PURE__ */ I(0);
	get first() {
		return W(this.#t);
	}
	set first(e) {
		L(this.#t, e, !0);
	}
	#n = /* @__PURE__ */ I(!1);
	get loadingOlder() {
		return W(this.#n);
	}
	set loadingOlder(e) {
		L(this.#n, e, !0);
	}
	#r = /* @__PURE__ */ I("");
	get error() {
		return W(this.#r);
	}
	set error(e) {
		L(this.#r, e, !0);
	}
	#i = /* @__PURE__ */ I(0);
	get version() {
		return W(this.#i);
	}
	set version(e) {
		L(this.#i, e, !0);
	}
	#a = new gi();
	#o = null;
	get blocks() {
		return this.#a.blocks;
	}
	get busy() {
		return this.#a.busy;
	}
	set busy(e) {
		this.#a.busy = e;
	}
	get usage() {
		return this.#a.usage;
	}
	constructor(e, t) {
		this.api = e, this.id = t;
	}
	open(e = 200) {
		this.close(), this.#o = this.api.session.stream(this.id, { tail: e }, (e) => {
			e.type === "open" ? this.status = "live" : e.type === "end" ? (this.status = "closed", this.busy = !1) : e.type === "error" ? this.status !== "closed" && (this.status = "reconnecting") : e.type === "event" && this.#s(e.event);
		});
	}
	close() {
		this.#o?.(), this.#o = null;
	}
	#s(e) {
		e.seq && (!this.first || e.seq < this.first) && (this.first = e.seq), this.#a.apply(e) && this.version++;
	}
	get earlier() {
		return this.first > 1 ? this.first - 1 : 0;
	}
	async loadOlder() {
		if (!this.loadingOlder && this.earlier) {
			this.loadingOlder = !0, this.error = "";
			try {
				let e = await this.api.session.older(this.id, this.first, 400), t = new gi();
				for (let n of e.events || []) t.apply(n);
				t.settle(), this.#a.blocks = [...ze(t.blocks), ...this.#a.blocks], this.first = Math.max(1, e.first || 1), this.version++;
			} catch (e) {
				this.error = String(e.message || e);
			} finally {
				this.loadingOlder = !1;
			}
		}
	}
	prompts() {
		return this.blocks.filter((e) => e.kind === "user" && e.source !== "controller").map((e) => e.text);
	}
};
// node_modules/marked/lib/marked.esm.js
function vi() {
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
var yi = vi();
function bi(e) {
	yi = e;
}
var xi = { exec: () => null };
function Si(e) {
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
			return i = i.replace(wi.caret, "$1"), n = n.replace(e, i), r;
		},
		getRegex: () => new RegExp(n, t)
	};
	return r;
}
var Ci = ((e = "") => {
	try {
		return !!RegExp("(?<=1)(?<!1)" + e);
	} catch {
		return !1;
	}
})(), wi = {
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
	nextBulletRegex: Si((e) => RegExp(`^ {0,${e}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`)),
	hrRegex: Si((e) => RegExp(`^ {0,${e}}((?:-[ 	]*){3,}|(?:_[ 	]*){3,}|(?:\\*[ 	]*){3,})(?:\\n+|$)`)),
	fencesBeginRegex: Si((e) => RegExp(`^ {0,${e}}(?:\`\`\`|~~~)`)),
	headingBeginRegex: Si((e) => RegExp(`^ {0,${e}}#{1,6}(?:\\s|$)`)),
	pedanticHeadingBeginRegex: Si((e) => RegExp(`^ {0,${e}}#`)),
	htmlBeginRegex: Si((e) => RegExp(`^ {0,${e}}(?:</?(?:${zi})(?: +|$|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)`, "i")),
	blockquoteBeginRegex: Si((e) => RegExp(`^ {0,${e}}>`))
}, Ti = /^(?:[ \t]*(?:\n|$))+/, Ei = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, Di = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}(?:(?=`)\1`*|(?=~)\1~*) *(?=\n|$)|$)/, Oi = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, ki = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, Ai = / {0,3}(?:[*+-]|\d{1,9}[.)])/, ji = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |fences|blockquote|heading|hr|html|table))+?)\n {0,3}(=+|-+)[ \t]*(?:\n+|$)/, Mi = Y(ji).replace(/bull/g, Ai).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), Ni = Y(ji).replace(/bull/g, Ai).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), Pi = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table|[ \t]+\n)[^\n]+)*)/, Fi = /^[^\n]+/, Ii = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\\n]|\n(?![ \t]*\n))+/, Li = Y(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", Ii).replace("title", /(?:"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\\n]|\n(?!\n))*'|\((?:\\[\s\S]|[^()\\])*\))/).getRegex(), Ri = Y(/^(bull)([ \t][^\n]*?)?(?:\n|$)/).replace(/bull/g, Ai).getRegex(), zi = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", Bi = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, Vi = Y("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n*|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>[^\\n]*\\n*|$)|<![A-Z][\\s\\S]*?(?:>[^\\n]*\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>[^\\n]*\\n*|$)|</?(tag)(?:[ \\t]+|\\n|/?>|$)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!(?:script|pre|style|textarea)(?![a-z0-9-]))([a-z][a-z0-9-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</[a-z][a-z0-9-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", Bi).replace("tag", zi).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), Hi = (e) => Y(Pi).replace("hr", Oi).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", e).replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", zi).getRegex(), Ui = Hi(/ {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]/), Wi = Hi(/ {0,3}(?:[*+-]|\d{1,9}[.)])(?:[ \t]|\n|$)/), Gi = {
	blockquote: Y(/^( {0,3}> ?(?:(?![ \t]*(?:\n|$))(paragraph)|[^\n]*)(?:\n|$))+/).replace("paragraph", Wi).getRegex(),
	code: Ei,
	def: Li,
	fences: Di,
	heading: ki,
	hr: Oi,
	html: Vi,
	lheading: Mi,
	list: Ri,
	newline: Ti,
	paragraph: Ui,
	table: xi,
	text: Fi
}, Ki = Y("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", Oi).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", zi).getRegex(), qi = {
	...Gi,
	lheading: Ni,
	table: Ki,
	paragraph: Y(Pi).replace("hr", Oi).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", Ki).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]+[^ \\t\\n]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", zi).getRegex()
}, Ji = {
	...Gi,
	html: Y("^ *(?:comment(?: *\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1>(?: *\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?>(?: *\\n{2,}|\\s*$))").replace("comment", Bi).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: xi,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: Y(Pi).replace("hr", Oi).replace("heading", " *#{1,6} *[^\n]").replace("lheading", Mi).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, Yi = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, Xi = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, Zi = /^( {2,}|\\)\n(?!\s*$)[ \t]*/, Qi = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, $i = /[\p{P}\p{S}]/u, ea = /[\s\p{P}\p{S}]/u, ta = /[^\s\p{P}\p{S}]/u, na = Y(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, ea).getRegex(), ra = /[\p{Pi}\p{Ps}"']/u, ia = /(?!~)[\p{P}\p{S}]/u, aa = /(?!~)[\s\p{P}\p{S}]/u, oa = /(?:[^\s\p{P}\p{S}]|~)/u, sa = Y(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", Ci ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), ca = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, la = Y(ca, "u").replace(/punct/g, $i).getRegex(), ua = Y(ca, "u").replace(/punct/g, ia).getRegex(), da = Y(/^(?:\*+(?:((?!\*)(?!openQuote)punct)|([^\s*]))?)|^_+(?:((?!_)(?!openQuote)punct)|([^\s_]))?/, "u").replace(/openQuote/g, ra).replace(/punct/g, $i).getRegex(), fa = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", pa = Y(fa, "gu").replace(/notPunctSpace/g, ta).replace(/punctSpace/g, ea).replace(/punct/g, $i).getRegex(), ma = Y(fa, "gu").replace(/notPunctSpace/g, oa).replace(/punctSpace/g, aa).replace(/punct/g, ia).getRegex(), ha = Y("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)[\\s](\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|(?:(?!\\*)punct|notPunctSpace)(\\*+)(?!\\*)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, ta).replace(/punctSpace/g, ea).replace(/punct/g, $i).getRegex(), ga = Y("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, ta).replace(/punctSpace/g, ea).replace(/punct/g, $i).getRegex(), _a = Y("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)[\\s](_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)|(?:(?!_)punct|notPunctSpace)(_+)(?!_)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, ta).replace(/punctSpace/g, ea).replace(/punct/g, $i).getRegex(), va = Y(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, $i).getRegex(), ya = Y("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, ta).replace(/punctSpace/g, ea).replace(/punct/g, $i).getRegex(), ba = Y(/\\(punct)/, "gu").replace(/punct/g, $i).getRegex(), xa = Y(/(`+)(?:[^`]|[^`][\s\S]*?[^`])\1(?!`)|`+|\\(punct)/, "gu").replace(/punct/g, $i).getRegex(), Sa = Y(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), Ca = Y(Bi).replace("(?:-->|$)", "-->").getRegex(), wa = Y("^comment|^</[a-zA-Z][a-zA-Z0-9-]*\\s*>|^<[a-zA-Z][a-zA-Z0-9-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", Ca).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), Ta = /\[(?:\\[\s\S]|[^\[\]\\])*\]/, Ea = Y(/(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/).replace("brackets", Ta).getRegex(), Da = Y(/^!?\[(label)\]\([ \t\n]*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?[ \t\n]*\)/).replace("label", Ea).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]+|(?=\))/).replace("title", /"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|\((?:\\[\s\S]|[^)\\])*\)/).getRegex(), Oa = Y(/^!?\[(label)\]\[(ref)\]/).replace("label", Ea).replace("ref", Ii).getRegex(), ka = Y(/^!?\[(ref)\](?:\[\])?/).replace("ref", Ii).getRegex(), Aa = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\]){1,999}/, ja = Y(/(?:[^\[\]\\`]*(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\]))){0,999}?[^\[\]\\`]*?/).replace("brackets", Ta).getRegex(), Ma = Y("reflink|nolink(?!\\()", "g").replace("reflink", Y(/^!?\[(label)\]\[(ref)\]/).replace("label", ja).replace("ref", Aa).getRegex()).replace("nolink", Y(/^!?\[(ref)\](?:\[\])?/).replace("ref", Aa).getRegex()).getRegex(), Na = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, Pa = Y(/(?:mailto:email|xmpp:email(?:\/[A-Za-z0-9@.]+)?)/).replace(/email/g, /[A-Za-z0-9._+-]+@[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(), Fa = {
	_backpedal: xi,
	anyPunctuation: ba,
	anyPunctuationOutsideCode: xa,
	autolink: Sa,
	blockSkip: sa,
	br: Zi,
	code: Xi,
	del: xi,
	delLDelim: xi,
	delRDelim: xi,
	emStrongLDelim: la,
	emStrongRDelimAst: pa,
	emStrongRDelimUnd: ga,
	escape: Yi,
	link: Da,
	nolink: ka,
	punctuation: na,
	reflink: Oa,
	reflinkSearch: Ma,
	tag: wa,
	text: Qi,
	url: xi
}, Ia = {
	...Fa,
	emStrongLDelim: da,
	emStrongRDelimAst: ha,
	emStrongRDelimUnd: _a,
	link: Y(/^!?\[(label)\]\((.*?)\)/).replace("label", Ea).getRegex(),
	reflink: Y(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", Ea).getRegex()
}, La = {
	...Fa,
	emStrongRDelimAst: ma,
	emStrongLDelim: ua,
	delLDelim: va,
	delRDelim: ya,
	url: Y(/^emailProtocol|^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("emailProtocol", Pa).replace("protocol", Na).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: Y(/^(?:[^a-zA-Z0-9](?=emailProtocol)|(`+|~+|[^`~])(?:(?=[`~])|(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|wwwStart|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9](?=emailProtocol)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))))/).replace("protocol", Na).replace(/emailProtocol/g, /(?:mailto|xmpp):/).replace("wwwStart", Ci ? "(?<![^\\s*_~(])www\\." : "www\\.").getRegex()
}, Ra = {
	...La,
	br: Y(Zi).replace("{2,}", "*").getRegex(),
	text: Y(La.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, za = {
	normal: Gi,
	gfm: qi,
	pedantic: Ji
}, Ba = {
	normal: Fa,
	gfm: La,
	breaks: Ra,
	pedantic: Ia
}, Va = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, Ha = (e) => Va[e];
function Ua(e, t) {
	if (t) {
		if (wi.escapeTest.test(e)) return e.replace(wi.escapeReplace, Ha);
	} else if (wi.escapeTestNoEncode.test(e)) return e.replace(wi.escapeReplaceNoEncode, Ha);
	return e;
}
function Wa(e) {
	return wi.numericCharacterReference.test(e) ? e.replace(wi.numericCharacterReferenceOrAmpersand, (e, t, n) => {
		if (t === void 0 && n === void 0) return "&amp;";
		let r = t === void 0 ? Number.parseInt(n, 16) : Number.parseInt(t, 10);
		return r === 0 || r > 1114111 || r >= 55296 && r <= 57343 ? "�" : r === 38 ? "&amp;" : String.fromCodePoint(r);
	}) : e;
}
function Ga(e) {
	try {
		e = encodeURI(e).replace(wi.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function Ka(e, t) {
	let n = e.replace(wi.findPipe, (e, t, n) => {
		let r = !1, i = t;
		for (; --i >= 0 && n[i] === "\\";) r = !r;
		return r ? "|" : " |";
	}).split(wi.splitPipe), r = 0;
	if (n[0].trim() || n.shift(), n.length > 0 && !n.at(-1)?.trim() && n.pop(), t) {
		if (n.length > t) n.splice(t);
		else for (; n.length < t;) n.push("");
	}
	for (; r < n.length; r++) n[r] = n[r].trim().replace(wi.slashPipe, "|");
	return n;
}
function qa(e, t, n) {
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
function Ja(e) {
	let t = e.split("\n"), n = t.length - 1;
	for (; n >= 0 && wi.blankLine.test(t[n]);) n--;
	return t.length - n <= 2 ? e : t.slice(0, n + 1).join("\n");
}
function Ya(e) {
	return e.trim().replace(wi.multipleSpaceGlobal, " ").toLowerCase().toUpperCase().toLowerCase();
}
function Xa(e, t) {
	if (e.indexOf(t[0]) === -1 && e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
var Za = /* @__PURE__ */ new WeakMap();
function Qa(e) {
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
function $a(e, t) {
	let n = Qa(t), r = Za.get(e);
	r ? r.push(n) : Za.set(e, [n]);
}
function eo(e) {
	let t = Za.get(e);
	t && (t.pop(), t.length === 0 && Za.delete(e));
}
function to(e, t, n) {
	let r = Za.get(e)?.at(-1);
	return r && r.maskedSrc === t ? r.lastPossibleEmStrongCloserIndex[n] : Qa(t).lastPossibleEmStrongCloserIndex[n];
}
function no(e, t) {
	let n = Za.get(e)?.at(-1);
	return n && n.maskedSrc === t ? n.lastPossibleDelCloserIndex : Qa(t).lastPossibleDelCloserIndex;
}
function ro(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function io(e, t, n, r, i) {
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
function ao(e, t, n) {
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
function oo(e, t, n, r) {
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
var so = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || yi;
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
			let e = this.options.pedantic ? t[0] : Ja(t[0]);
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
			let e = t[0], n = ao(e, t[3] || "", this.rules);
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
				let t = qa(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceTabChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: qa(t[0], "\n"),
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
			raw: qa(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = qa(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteMarkerWhitespace, (e, t, n) => t + ro(n, t.length)).replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
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
				let u = t[2].split("\n", 1)[0], d = t[1].length, f = this.options.pedantic ? ro(u, d) : u.replace(this.rules.other.leadingSpaceTab, (e) => ro(e, d)), p = e.split("\n", 1)[0], m = !f.trim(), h = 0;
				if (this.options.pedantic ? (h = 2, a = f.trimStart()) : m ? h = d + 1 : (h = f.search(this.rules.other.nonSpaceChar), h = h > 4 ? 1 : h, a = f.slice(h), h += d), m && this.rules.other.blankLine.test(p) && (e = e.substring(p.length + 1), i = !0), !i) {
					let t = this.rules.other.nextBulletRegex(h), n = this.rules.other.hrRegex(h), r = this.rules.other.fencesBeginRegex(h), i = this.options.pedantic ? this.rules.other.pedanticHeadingBeginRegex(h) : this.rules.other.headingBeginRegex(h), o = this.rules.other.htmlBeginRegex(h), s = this.rules.other.blockquoteBeginRegex(h);
					for (; e;) {
						let c = e.split("\n", 1)[0], l;
						if (p = c, this.options.pedantic ? (p = p.replace(this.rules.other.listReplaceNesting, "  "), l = p) : l = p.replace(this.rules.other.leadingSpaceTab, (e) => ro(e)), r.test(p) || i.test(p) || o.test(p) || s.test(p) || t.test(p) || n.test(p)) break;
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
			let e = Ja(t[0]);
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
			if (!this.rules.other.startAngleBracket.test(t[2]) && Xa(t[2], "()") !== -1) return;
			let e = Ya(t[1]), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: qa(t[0], "\n"),
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = Ka(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: qa(t[0], "\n"),
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
			for (let e of i) a.rows.push(Ka(e, a.header.length).map((e, t) => ({
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
				raw: qa(t[0], "\n"),
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
			if (!this.options.pedantic && oo(e, t[1], n, this.rules)) return;
			let r = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(r)) {
				if (!this.rules.other.endAngleBracket.test(r)) return;
				let e = qa(r.slice(0, -1), "\\");
				if ((r.length - e.length) % 2 == 0) return;
			} else {
				let e = Xa(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let i = qa(t[0].slice(0, -1), "\\");
			if ((t[0].length - i.length) % 2 == 0) return;
			let a = t[2], o = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(a);
				e && (a = e[1], o = e[3]);
			} else o = t[3] ? t[3].slice(1, -1) : "";
			return a = a.trim(), this.rules.other.startAngleBracket.test(a) && (a = this.options.pedantic && !this.rules.other.endAngleBracket.test(r) ? a.slice(1) : a.slice(1, -1)), io(t, {
				href: a && a.replace(this.rules.inline.anyPunctuation, "$1"),
				title: o && o.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let r = n[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && oo(e, n[1], r, this.rules)) return;
			let i = t[Ya(n[2] || n[1])];
			if (!i) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return io(n, i, n[0], this.lexer, this.rules);
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
			let _ = i === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd, v = to(this, t, i), y = t.length - e.length + n;
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
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim, c = no(this, t), l = t.length - e.length + n;
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
				text: e ? t[0] : Wa(t[0]),
				escaped: e
			};
		}
	}
}, co = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || yi, this.options.tokenizer = this.options.tokenizer || new so(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			linkEmitted: !1,
			linkParenPossible: !0,
			top: !0
		};
		let t = {
			other: wi,
			block: za.normal,
			inline: Ba.normal
		};
		this.options.pedantic ? (t.block = za.pedantic, t.inline = Ba.pedantic) : this.options.gfm && (t.block = za.gfm, t.inline = this.options.breaks ? Ba.breaks : Ba.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: za,
			inline: Ba
		};
	}
	static lex(t, n) {
		return new e(n).lex(t);
	}
	static lexInline(t, n) {
		return new e(n).inlineTokens(t);
	}
	lex(e) {
		e = e.replace(wi.carriageReturn, "\n"), this.blockTokens(e, this.tokens);
		for (let e = 0; e < this.inlineQueue.length; e++) {
			let t = this.inlineQueue[e];
			this.state.inRawBlock = !1, this.inlineTokens(t.src, t.tokens);
		}
		return this.inlineQueue = [], this.tokens;
	}
	blockTokens(e, t = [], n = !1) {
		this.tokenizer.lexer = this, this.options.pedantic && (e = e.replace(wi.tabCharGlobal, "    ").replace(wi.spaceLine, ""));
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
			if (e.charAt(0) !== "!" && Object.hasOwn(this.tokens.links, Ya(e.slice(n + 1, -1))) && !(n > 1 && this.linkInText(e.slice(1, n - 1)))) return !0;
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
		let n = e.includes("*") || e.includes("_") || e.includes("~") || this.options.hooks?.emStrongMask || this.tokenizer.emStrong !== so.prototype.emStrong || this.tokenizer.del !== so.prototype.del ? this.#n(e) : e;
		$a(this.tokenizer, n);
		try {
			return this.#t(e, n, t);
		} finally {
			eo(this.tokenizer);
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
				if (!Object.hasOwn(this.tokens.links, Ya(t.slice(r + 1, -1)))) return t;
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
}, lo = class {
	options;
	parser;
	constructor(e) {
		this.options = e || yi;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(wi.notSpaceStart)?.[0], i = e ? e.replace(wi.endingNewline, "") + "\n" : "";
		return r ? "<pre><code class=\"language-" + Ua(r) + "\">" + (n ? i : Ua(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : Ua(i, !0)) + "</code></pre>\n";
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
		return `<code>${Ua(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, text: n, tokens: r, autolink: i }) {
		let a = i ? Ua(n, !0) : this.parser.parseInline(r), o = Ga(e);
		if (o === null) return a;
		e = Ua(o, i);
		let s = "<a href=\"" + e + "\"";
		return t && (s += " title=\"" + Ua(t) + "\""), s += ">" + a + "</a>", s;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = Ga(e);
		if (i === null) return Ua(n);
		e = i;
		let a = `<img src="${Ua(e)}" alt="${Ua(n)}"`;
		return t && (a += ` title="${Ua(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : Ua(e.text);
	}
}, uo = class {
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
}, fo = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || yi, this.options.renderer = this.options.renderer || new lo(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new uo(), this.textRenderer.parser = this;
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
}, po = class {
	options;
	block;
	constructor(e) {
		this.options = e || yi;
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
		return e ? co.lex : co.lexInline;
	}
	provideParser(e = this.block) {
		return e ? fo.parse : fo.parseInline;
	}
};
function mo(e, t) {
	if (Array.isArray(t)) for (let n = 0; n < t.length; n++) e.push(t[n]);
	else e.push(t);
}
var ho = new class {
	defaults = vi();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = fo;
	Renderer = lo;
	TextRenderer = uo;
	Lexer = co;
	Tokenizer = so;
	Hooks = po;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (mo(n, t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) mo(n, this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) mo(n, this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				mo(n, this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					mo(n, this.walkTokens(i, t));
				}) : e.tokens && mo(n, this.walkTokens(e.tokens, t));
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
				let t = this.defaults.renderer || new lo(this.defaults);
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
				let t = this.defaults.tokenizer || new so(this.defaults);
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
				let t = this.defaults.hooks || new po();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = po.passThroughHooks.has(n) ? (e) => {
						if (t.options.async && po.passThroughHooksRespectAsync.has(n)) return (async () => {
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
					return mo(n, r.call(this, e)), t && mo(n, t.call(this, e)), n;
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
		return co.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return fo.parse(e, t ?? this.defaults);
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
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? co.lex : co.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? fo.parse : fo.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? co.lex : co.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? fo.parse : fo.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + Ua(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function X(e, t) {
	return ho.parse(e, t);
}
X.options = X.setOptions = function(e) {
	return ho.setOptions(e), X.defaults = ho.defaults, bi(X.defaults), X;
}, X.getDefaults = vi, X.defaults = yi;
function go(...e) {
	return ho.use(...e), X.defaults = ho.defaults, bi(X.defaults), X;
}
X.use = go, X.walkTokens = function(e, t) {
	return ho.walkTokens(e, t);
}, X.parseInline = ho.parseInline, X.Parser = fo, X.parser = fo.parse, X.Renderer = lo, X.TextRenderer = uo, X.Lexer = co, X.lexer = co.lex, X.Tokenizer = so, X.Hooks = po, X.parse = X, X.options, X.setOptions, X.walkTokens, X.parseInline, fo.parse, co.lex;
// node_modules/dompurify/dist/purify.es.mjs
function _o(e, t) {
	this.v = e, this.k = t;
}
function vo(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function yo(e) {
	if (Array.isArray(e)) return e;
}
function bo(e, t) {
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
function xo() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function So(e, t) {
	return yo(e) || bo(e, t) || Co(e, t) || xo();
}
function Co(e, t) {
	if (e) {
		if (typeof e == "string") return vo(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? vo(e, t) : void 0;
	}
}
function wo(e) {
	var t, n;
	function r(t, n) {
		try {
			var a = e[t](n), o = a.value, s = o instanceof _o;
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
wo.prototype[typeof Symbol == "function" && Symbol.asyncIterator || "@@asyncIterator"] = function() {
	return this;
}, wo.prototype.next = function(e) {
	return this._invoke("next", e);
}, wo.prototype.throw = function(e) {
	return this._invoke("throw", e);
}, wo.prototype.return = function(e) {
	return this._invoke("return", e);
};
var To = Object.entries, Eo = Object.setPrototypeOf, Do = Object.isFrozen, Oo = Object.getPrototypeOf, ko = Object.getOwnPropertyDescriptor, Ao = Object.freeze, jo = Object.seal, Mo = Object.create, No = typeof Reflect < "u" && Reflect, Po = No.apply, Fo = No.construct;
Ao ||= function(e) {
	return e;
}, jo ||= function(e) {
	return e;
}, Po ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, Fo ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var Io = Z(Array.prototype.forEach);
Array.prototype.indexOf;
var Lo = Z(Array.prototype.lastIndexOf), Ro = Z(Array.prototype.pop), zo = Z(Array.prototype.push), Bo = Z(Array.prototype.slice), Vo = Z(Array.prototype.splice), Ho = Array.isArray, Uo = os(Set), Wo = Z(Set.prototype.add), Go = Z(Set.prototype.has), Ko = Z(String.prototype.toLowerCase), qo = Z(String.prototype.toString), Jo = Z(String.prototype.match), Yo = Z(String.prototype.replace), Xo = Z(String.prototype.indexOf), Zo = Z(String.prototype.trim), Qo = Z(Number.prototype.toString), $o = Z(Boolean.prototype.toString), es = typeof BigInt > "u" ? null : Z(BigInt.prototype.toString), ts = typeof Symbol > "u" ? null : Z(Symbol.prototype.toString), ns = Z(Object.prototype.hasOwnProperty), rs = Z(Object.prototype.toString), is = Z(RegExp.prototype.test), as = os(TypeError);
function Z(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return Po(e, t, n);
	};
}
function os(e) {
	return function() {
		return Fo(e, [...arguments]);
	};
}
function Q(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : Ko;
	if (Eo && Eo(e, null), !Ho(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (Do(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function ss(e) {
	for (let t = 0; t < e.length; t++) ns(e, t) || (e[t] = null);
	return e;
}
function cs(e) {
	let t = Mo(null);
	for (let r of To(e)) {
		var n = So(r, 2);
		let i = n[0], a = n[1];
		ns(e, i) && (t[i] = Ho(a) ? ss(a) : a && typeof a == "object" && a.constructor === Object ? cs(a) : a);
	}
	return t;
}
function ls(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return Qo(e);
		case "boolean": return $o(e);
		case "bigint": return es ? es(e) : "0";
		case "symbol": return ts ? ts(e) : "Symbol()";
		case "undefined": return rs(e);
		case "function":
		case "object": {
			if (e === null) return rs(e);
			let t = e, n = us(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : rs(e);
			}
			return rs(e);
		}
		default: return rs(e);
	}
}
function us(e, t) {
	for (; e !== null;) {
		let n = ko(e, t);
		if (n) {
			if (n.get) return Z(n.get);
			if (typeof n.value == "function") return Z(n.value);
		}
		e = Oo(e);
	}
	function n() {
		return null;
	}
	return n;
}
function ds(e, t) {
	for (; e !== null;) {
		let n = ko(e, t);
		if (n) return n.get ? Z(n.get) : null;
		e = Oo(e);
	}
	return null;
}
function fs(e) {
	try {
		return is(e, ""), !0;
	} catch {
		return !1;
	}
}
var ps = Ao(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), ms = Ao(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), hs = Ao([
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
]), gs = Ao([
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
]), _s = Ao(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), vs = Ao([
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
]), ys = Ao(["#text"]), bs = Ao(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), xs = Ao(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dominant-baseline.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.pointer-events.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-orientation.text-rendering.textlength.type.u1.u2.unicode.values.vector-effect.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), Ss = Ao(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), Cs = Ao([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), ws = jo(/{{[\w\W]*|^[\w\W]*}}/g), Ts = jo(/<%[\w\W]*|^[\w\W]*%>/g), Es = jo(/\${[\w\W]*/g), Ds = jo(/^data-[\-\w.\u00B7-\uFFFF]+$/), Os = jo(/^aria-[\-\w]+$/), ks = jo(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), As = jo(/^(?:\w+script|data):/i), js = jo(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), Ms = jo(/^html$/i), Ns = jo(/^[a-z][.\w]*(-[.\w]+)+$/i), Ps = jo(/<[/\w!]/g), Fs = jo(/<[/\w]/g), Is = jo(/<\/no(script|embed|frames)/i), Ls = jo(/\/>/i), $ = {
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
}, Rs = Ao(Q({}, [
	"style",
	"script",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"plaintext",
	"noscript"
])), zs = function() {
	return typeof window > "u" ? null : window;
}, Bs = function(e, t) {
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
}, Vs = function() {
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
}, Hs = function(e, t, n, r) {
	return ns(e, t) && Ho(e[t]) ? Q(r.base ? cs(r.base) : {}, e[t], r.transform) : n;
}, Us = function(e, t, n) {
	let r = ns(e, t) ? e[t] : void 0;
	return r && typeof r == "object" ? cs(r) : n();
};
function Ws() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : zs(), t = (e) => Ws(e);
	if (t.version = "3.4.17", t.removed = [], !e || !e.document || e.document.nodeType !== $.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = us(d, "cloneNode"), p = us(d, "remove"), m = us(d, "removeAttributeNode"), h = us(d, "nextSibling"), g = us(d, "childNodes"), _ = us(d, "parentNode"), v = us(d, "shadowRoot"), y = us(d, "attributes"), b = o && o.prototype ? us(o.prototype, "nodeType") : null, x = o && o.prototype ? us(o.prototype, "nodeName") : null, S = o && o.prototype ? us(o.prototype, "ownerDocument") : null, C = function(e) {
		return b ? b(e) : e.nodeType;
	}, w = function(e) {
		return x ? x(e) : e.nodeName;
	}, ee = ds(d, "namespaceURI"), te = function(e) {
		return ee ? ee(e) : e.namespaceURI;
	};
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let T, E = "", ne, re = !1, ie = 0, ae = function() {
		if (ie > 0) throw as("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
	}, oe = function(e) {
		ae(), ie++;
		try {
			return T.createHTML(e);
		} finally {
			ie--;
		}
	}, se = function(e) {
		ae(), ie++;
		try {
			return T.createScriptURL(e);
		} finally {
			ie--;
		}
	}, ce = function() {
		return re ||= (ne = Bs(u, i), !0), ne;
	}, le = n, ue = le.implementation, de = le.createNodeIterator, fe = le.createDocumentFragment, pe = le.getElementsByTagName, me = r.importNode, D = Vs();
	t.isSupported = typeof To == "function" && typeof _ == "function" && ue && ue.createHTMLDocument !== void 0;
	let he = ws, ge = Ts, _e = Es, ve = Ds, ye = Os, O = As, be = js, k = Ns, xe = ks, A = null, j = Q({}, [
		...ps,
		...ms,
		...hs,
		..._s,
		...ys
	]), M = null, Se = Q({}, [
		...bs,
		...xs,
		...Ss,
		...Cs
	]), Ce = Object.seal(Mo(null, {
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
	})), we = null, Te = null, Ee = Object.seal(Mo(null, {
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
	})), De = !0, Oe = !0, ke = !1, Ae = !0, je = !1, Me = !0, Ne = !1, Pe = !1, Fe = null, Ie = null, Le = !1, Re = !1, ze = !1, Be = !1, Ve = !0, He = !1, Ue = "user-content-", We = !0, Ge = !1, Ke = null, qe = {}, Je = null, Ye = Q({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Xe = null, Ze = Q({}, [
		"audio",
		"video",
		"img",
		"source",
		"image",
		"track"
	]), N = null, Qe = Q({}, [
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
	]), $e = "http://www.w3.org/1998/Math/MathML", et = "http://www.w3.org/2000/svg", tt = "http://www.w3.org/1999/xhtml", nt = tt, rt = !1, it = null, at = Q({}, [
		$e,
		et,
		tt
	], qo), ot = Ao([
		"mi",
		"mo",
		"mn",
		"ms",
		"mtext"
	]), st = Q({}, ot), ct = Ao(["annotation-xml"]), lt = Q({}, ct), ut = Q({}, [
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
		e && typeof e == "object" || (e = {}), e = cs(e), dt = ft.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, P = dt === "application/xhtml+xml" ? qo : Ko, A = Hs(e, "ALLOWED_TAGS", j, { transform: P }), M = Hs(e, "ALLOWED_ATTR", Se, { transform: P }), it = Hs(e, "ALLOWED_NAMESPACES", at, { transform: qo }), N = Hs(e, "ADD_URI_SAFE_ATTR", Qe, {
			transform: P,
			base: Qe
		}), Xe = Hs(e, "ADD_DATA_URI_TAGS", Ze, {
			transform: P,
			base: Ze
		}), Je = Hs(e, "FORBID_CONTENTS", Ye, { transform: P }), we = Hs(e, "FORBID_TAGS", cs({}), { transform: P }), Te = Hs(e, "FORBID_ATTR", cs({}), { transform: P }), qe = ns(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? cs(e.USE_PROFILES) : e.USE_PROFILES : !1, De = e.ALLOW_ARIA_ATTR !== !1, Oe = e.ALLOW_DATA_ATTR !== !1, ke = e.ALLOW_UNKNOWN_PROTOCOLS || !1, Ae = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, je = e.SAFE_FOR_TEMPLATES || !1, Me = e.SAFE_FOR_XML !== !1, Ne = e.WHOLE_DOCUMENT || !1, Re = e.RETURN_DOM || !1, ze = e.RETURN_DOM_FRAGMENT || !1, Be = e.RETURN_TRUSTED_TYPE || !1, Le = e.FORCE_BODY || !1, Ve = e.SANITIZE_DOM !== !1, He = e.SANITIZE_NAMED_PROPS || !1, We = e.KEEP_CONTENT !== !1, Ge = e.IN_PLACE || !1, xe = fs(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : ks, nt = typeof e.NAMESPACE == "string" ? e.NAMESPACE : tt, st = Us(e, "MATHML_TEXT_INTEGRATION_POINTS", () => Q({}, ot)), lt = Us(e, "HTML_INTEGRATION_POINTS", () => Q({}, ct));
		let t = Us(e, "CUSTOM_ELEMENT_HANDLING", () => Mo(null));
		if (Ce = Mo(null), ns(t, "tagNameCheck") && ht(t.tagNameCheck) && (Ce.tagNameCheck = t.tagNameCheck), ns(t, "attributeNameCheck") && ht(t.attributeNameCheck) && (Ce.attributeNameCheck = t.attributeNameCheck), ns(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (Ce.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), jo(Ce), je && (Oe = !1), ze && (Re = !0), qe && (A = Q({}, ys), M = Mo(null), qe.html === !0 && (Q(A, ps), Q(M, bs)), qe.svg === !0 && (Q(A, ms), Q(M, xs), Q(M, Cs)), qe.svgFilters === !0 && (Q(A, hs), Q(M, xs), Q(M, Cs)), qe.mathMl === !0 && (Q(A, _s), Q(M, Ss), Q(M, Cs))), Ee.tagCheck = null, Ee.attributeCheck = null, ns(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? Ee.tagCheck = e.ADD_TAGS : Ho(e.ADD_TAGS) && (A === j && (A = cs(A)), Q(A, e.ADD_TAGS, P))), ns(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? Ee.attributeCheck = e.ADD_ATTR : Ho(e.ADD_ATTR) && (M === Se && (M = cs(M)), Q(M, e.ADD_ATTR, P))), ns(e, "ADD_FORBID_CONTENTS") && Ho(e.ADD_FORBID_CONTENTS) && (Je === Ye && (Je = cs(Je)), Q(Je, e.ADD_FORBID_CONTENTS, P)), We && (A["#text"] = !0), Ne && Q(A, [
			"html",
			"head",
			"body"
		]), A.table && (Q(A, ["tbody"]), delete we.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw as("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw as("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
			let t = T;
			T = e.TRUSTED_TYPES_POLICY;
			try {
				E = oe("");
			} catch (e) {
				throw T = t, e;
			}
		} else e.TRUSTED_TYPES_POLICY === null ? (T = void 0, E = "") : (T === void 0 && (T = ce()), T && typeof E == "string" && (E = oe("")));
		Ao && Ao(e), pt = e;
	}, _t = Q({}, [
		...ms,
		...hs,
		...gs
	]), vt = Q({}, [..._s, ...vs]), F = function(e, t, n) {
		return t.namespaceURI === tt ? e === "svg" : t.namespaceURI === $e ? e === "svg" && (n === "annotation-xml" || st[n]) : !!_t[e];
	}, yt = function(e, t, n) {
		return t.namespaceURI === tt ? e === "math" : t.namespaceURI === et ? e === "math" && lt[n] : !!vt[e];
	}, bt = function(e, t, n) {
		return t.namespaceURI === et && !lt[n] || t.namespaceURI === $e && !st[n] ? !1 : !vt[e] && (ut[e] || !_t[e]);
	}, xt = function(e) {
		let t = te(e), n = Ko(w(e)), r = _(e), i, a;
		return r && C(r) === $.element ? (i = { namespaceURI: te(r) }, a = Ko(w(r))) : (i = { namespaceURI: nt }, a = "template"), it[t] ? t === et ? F(n, i, a) : t === $e ? yt(n, i, a) : t === tt ? bt(n, i, a) : !!(dt === "application/xhtml+xml" && it[t]) : !1;
	}, St = function(e) {
		zo(t.removed, { element: e });
		try {
			_(e).removeChild(e);
		} catch {
			if (p(e), !_(e)) throw as("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
		}
	}, Ct = function(e, t, n) {
		try {
			m(e, t);
		} catch {
			try {
				e.removeAttribute(n);
			} catch {}
		}
	}, wt = function(e) {
		Ot(e);
		let t = g(e);
		t && Io(Bo(t), (e) => {
			try {
				p(e);
			} catch {}
		});
		let n = C(e) === $.element ? y(e) : null;
		if (n) for (let t = n.length - 1; t >= 0; --t) {
			let r = n[t], i = r && r.name;
			typeof i == "string" && Ct(e, r, i);
		}
	}, Tt = function(e, n, r) {
		if (!r) try {
			r = n.getAttributeNode(e);
		} catch {
			r = null;
		}
		zo(t.removed, {
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
			if (Re || ze) try {
				St(n);
			} catch {}
			else try {
				n.setAttribute(e, "");
			} catch {}
		}
	}, Et = function(e) {
		let t = y(e);
		if (t) for (let n = t.length - 1; n >= 0; --n) {
			let r = t[n], i = r && r.name;
			typeof i != "string" || M[P(i)] || Ct(e, r, i);
		}
	}, Dt = function(e, t) {
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
		Dt(e, (e) => {
			C(e) === $.element && Et(e);
		});
	}, kt = function(e) {
		let t = [];
		return Dt(e, (e) => {
			t.push(e);
		}, !0), t;
	}, At = function(e, t) {
		let n = Uo();
		Io(kt(e), (e) => {
			C(e) === $.element && Wo(n, e);
		}), Io(t, (e) => {
			C(e) === $.element && !Go(n, e) && Et(e);
		});
	}, jt = function(e, n) {
		let r = !1;
		return Io(t.removed, (t) => {
			t.element && (r ||= t.element === e, Ot(t.element));
		}), At(e, n), r;
	}, Mt = function(e, t) {
		let n = Ke, r = null, i = () => (r || (r = Uo(), Io(t, (e) => {
			Wo(r, e);
		})), r);
		for (let t = 0;; t++) {
			let r = kt(e), a = [e], o = !1;
			if (Io(r, (e) => {
				C(e) === $.documentFragment ? zo(a, e) : !o && !Go(n, e) && (o = Go(i(), e));
			}), !o) return;
			if (t >= 8) throw as("the tree kept changing during in-place sanitization; refusing to return it");
			Io(a, (t, n) => {
				Nt(t, n > 0, e, i());
			});
		}
	}, Nt = function(e, t, n, r) {
		let i = Rt(e), a;
		for (; a = i.nextNode();) Go(r, a) && !Go(Ke, a) && (t && Ut(D.uponSanitizeShadowNode, a, null), Yt(a, n), R(a, n), Vt(a.content) && nn(a.content));
	}, Pt = function(e) {
		Ke && Wo(Ke, e);
	}, Ft = function(e, t) {
		return Me ? e === "patchsrc" || e === "for" && t !== "label" && t !== "output" : !1;
	}, It = function(e) {
		Me && Dt(e, (e) => {
			let t = C(e);
			if (t === $.processingInstruction || t === $.comment && is(Fs, e.data)) {
				try {
					p(e);
				} catch {}
				return !1;
			}
			if (t === $.element) {
				let t = e, n = P(w(e));
				try {
					t.hasAttribute && t.hasAttribute("patchsrc") && t.removeAttribute("patchsrc"), t.hasAttribute && t.hasAttribute("for") && Ft("for", n) && t.removeAttribute("for");
				} catch {}
			}
			return !0;
		});
	}, Lt = function(e) {
		let t = null, r = null;
		if (Le) e = "<remove></remove>" + e;
		else {
			let t = Jo(e, /^[\r\n\t ]+/);
			r = t && t[0];
		}
		dt === "application/xhtml+xml" && nt === tt && (e = "<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>" + e + "</body></html>");
		let i = T ? oe(e) : e;
		if (nt === tt) try {
			t = new l().parseFromString(i, dt);
		} catch {}
		if (!t || !t.documentElement) {
			t = ue.createDocument(nt, "template", null);
			try {
				t.documentElement.innerHTML = rt ? E : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), nt === tt ? pe.call(t, Ne ? "html" : "body")[0] : Ne ? t.documentElement : a;
	}, Rt = function(e) {
		let t = S ? S(e) : e.ownerDocument;
		return de.call(t || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, zt = function(e) {
		return e = Yo(e, he, " "), e = Yo(e, ge, " "), e = Yo(e, _e, " "), e;
	}, I = function(e) {
		Bt(e);
		let t = e.querySelectorAll?.call(e, "template");
		t && Io(t, (e) => {
			Vt(e.content) && I(e.content);
		});
	}, Bt = function(e) {
		e.normalize();
		let t = S ? S(e) : e.ownerDocument, n = de.call(t || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), r = n.nextNode();
		for (; r;) r.data = zt(r.data), r = n.nextNode();
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
	function Ut(e, n, r) {
		e.length !== 0 && Io(e, (e) => {
			e.call(t, n, r, pt);
		});
	}
	let Wt = function(e, t) {
		return !!(Me && e.hasChildNodes() && !Ht(e.firstElementChild) && is(Ps, e.textContent) && is(Ps, e.innerHTML) || Me && Rs[t] && C(e) === $.element && te(e) === tt && (Ht(e.firstElementChild) || typeof e.textContent == "string" && is(Ps, e.textContent)) || e.nodeType === $.processingInstruction || Me && e.nodeType === $.comment && is(Fs, e.data));
	}, Gt = function(e, t) {
		return e instanceof RegExp ? is(e, t) : e instanceof Function && !!e(t, ...[...arguments].slice(2));
	}, Kt = function(e, t, n) {
		if (!we[t] && Qt(t) && Gt(Ce.tagNameCheck, t)) return !1;
		if (We && !Je[t] && !Je[Ko(t)]) {
			let t = _(e), r = g(e);
			if (r && t) {
				let i = r.length;
				for (let a = i - 1; a >= 0; --a) {
					let i = e === n ? f(r[a], !0) : r[a];
					t.insertBefore(i, h(e));
				}
			}
		}
		return St(e), !0;
	}, qt = function(e, t, n, r) {
		return e.length === 0 ? t : t === n || t === r ? cs(t) : t;
	}, Jt = function(e, t) {
		return e === t || _(e) !== null ? !1 : (Ge && Ot(e), !0);
	}, Yt = function(e, n) {
		if (Pt(e), Ut(D.beforeSanitizeElements, e, null), Jt(e, n)) return !0;
		if (L(e)) return St(e), !0;
		let r = P(w(e));
		if (A = qt(D.uponSanitizeElement, A, j, Fe), Ut(D.uponSanitizeElement, e, {
			tagName: r,
			allowedTags: A
		}), Jt(e, n)) return !0;
		if (Wt(e, r)) return St(e), !0;
		if (we[r] || !(Ee.tagCheck instanceof Function && Ee.tagCheck(r)) && !A[r]) {
			let t = Kt(e, r, n);
			return t === !1 && (Ut(D.afterSanitizeElements, e, null), Jt(e, n)) ? !0 : t;
		}
		if (C(e) === $.element && !xt(e) || (r === "noscript" || r === "noembed" || r === "noframes") && is(Is, e.innerHTML)) return St(e), !0;
		if (je && e.nodeType === $.text) {
			let n = zt(e.textContent);
			e.textContent !== n && (zo(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return Ut(D.afterSanitizeElements, e, null), Jt(e, n);
	}, Xt = function(e, t, r) {
		if (Te[t] || Ft(t, e) || Ve && (t === "id" || t === "name") && (r in n || r in mt)) return !1;
		let i = M[t] || Ee.attributeCheck instanceof Function && Ee.attributeCheck(t, e);
		return Oe && is(ve, t) || De && is(ye, t) ? !0 : i ? N[t] || is(xe, Yo(r, be, "")) || (t === "src" || t === "xlink:href" || t === "href") && e !== "script" && Xo(r, "data:") === 0 && Xe[e] || ke && !is(O, Yo(r, be, "")) ? !0 : !r : Qt(e) && Gt(Ce.tagNameCheck, e) && Gt(Ce.attributeNameCheck, t, e) || t === "is" && Ce.allowCustomizedBuiltInElements && Gt(Ce.tagNameCheck, r);
	}, Zt = Q({}, [
		"annotation-xml",
		"color-profile",
		"font-face",
		"font-face-format",
		"font-face-name",
		"font-face-src",
		"font-face-uri",
		"missing-glyph"
	]), Qt = function(e) {
		return !Zt[Ko(e)] && is(k, e);
	}, $t = function(e, t, n, r) {
		if (T && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return oe(r);
			case "TrustedScriptURL": return se(r);
		}
		return r;
	}, en = function(e, t, n, r) {
		try {
			return n ? e.setAttributeNS(n, t, r) : e.setAttribute(t, r), !L(e) || (St(e), !1);
		} catch {
			return Tt(t, e), !1;
		}
	}, tn = function(e, t) {
		return Me && is(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, t) ? !0 : Ko(e) === "attributename" && Jo(t, "href") !== null;
	}, R = function(e, n) {
		if (Ut(D.beforeSanitizeAttributes, e, null), Jt(e, n)) return;
		let r = e.attributes;
		if (!r || L(e)) return;
		M = qt(D.uponSanitizeAttribute, M, Se, Ie);
		let i = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: M,
			forceKeepAttr: void 0
		}, a = r.length, o = P(e.nodeName);
		for (; a--;) {
			let n = r[a], s = n.name, c = n.namespaceURI, l = n.value, u = P(s), d = l, f = s === "value" ? d : Zo(d), p = !1;
			i.attrName = u, i.attrValue = f, i.keepAttr = !0, i.forceKeepAttr = void 0, Ut(D.uponSanitizeAttribute, e, i), f = i.attrValue, He && (u === "id" || u === "name") && Xo(f, Ue) !== 0 && (Tt(s, e, n), f = Ue + f, p = !0), tn(s, f) ? Tt(s, e, n) : i.forceKeepAttr || (i.keepAttr ? !Ae && is(Ls, f) ? Tt(s, e, n) : (je && (f = zt(f)), Xt(o, u, f) ? (f = $t(o, u, c, f), f !== d && en(e, s, c, f) && p && Ro(t.removed)) : Tt(s, e, n)) : Tt(s, e, n));
		}
		Ut(D.afterSanitizeAttributes, e, null), Jt(e, n);
	}, nn = function(e) {
		let t = null, n = Rt(e);
		for (Ut(D.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (Ut(D.uponSanitizeShadowNode, t, null), Yt(t, e), R(t, e), Vt(t.content) && nn(t.content), C(t) === $.element) {
			let e = v(t);
			Vt(e) && (z(e), nn(e));
		}
		je && Bt(e), Ut(D.afterSanitizeShadowDOM, e, null);
	}, z = function(e) {
		let t = [{
			node: e,
			shadow: null
		}];
		for (; t.length > 0;) {
			let e = t.pop();
			if (e.shadow) {
				nn(e.shadow);
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
		if (rt = !e, rt && (e = "<!-->"), typeof e != "string" && !Ht(e) && (e = ls(e), typeof e != "string")) throw as("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		Pe ? (A = Fe, M = Ie) : gt(n), (D.uponSanitizeElement.length > 0 || D.uponSanitizeAttribute.length > 0) && (A = cs(A)), D.uponSanitizeAttribute.length > 0 && (M = cs(M)), t.removed = [];
		let c = Ge && typeof e != "string" && Ht(e), l = [], u = Ke;
		if (c) {
			It(e);
			let t = w(e);
			if (typeof t == "string") {
				let n = P(t);
				if (!A[n] || we[n]) throw wt(e), as("root node is forbidden and cannot be sanitized in-place");
			}
			if (L(e)) throw wt(e), as("root node is clobbered and cannot be sanitized in-place");
			l = kt(e), Ke = Uo();
		} else if (Ht(e)) i = Lt("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === $.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), z(i);
		else {
			if (!Re && !je && !Ne && e.indexOf("<") === -1) return T && Be ? oe(e) : e;
			if (i = Lt(e), !i) return Re ? null : Be ? E : "";
		}
		i && Le && St(i.firstChild);
		let d = c ? e : i;
		try {
			c && z(e);
			let t = Rt(d);
			for (; o = t.nextNode();) Yt(o, d), R(o, d), Vt(o.content) && nn(o.content);
			c && Mt(e, l);
		} catch (t) {
			throw c && (wt(e), jt(e, l)), t;
		} finally {
			Ke = u;
		}
		if (c) {
			if (jt(e, l)) throw as("a node selected for removal could not be safely returned; refusing to sanitize in place");
			return je && I(e), e;
		}
		if (Re) {
			if (je && I(i), ze) for (s = fe.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (M.shadowroot || M.shadowrootmode) && (s = me.call(r, s, !0)), s;
		}
		let f = Ne ? i.outerHTML : i.innerHTML;
		return Ne && A["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && is(Ms, i.ownerDocument.doctype.name) && (f = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + f), je && (f = zt(f)), T && Be ? oe(f) : f;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		gt(e), Pe = !0, Fe = A, Ie = M;
	}, t.clearConfig = function() {
		pt = null, Pe = !1, Fe = null, Ie = null, T = ne, E = "";
	}, t.isValidAttribute = function(e, t, n) {
		pt || gt({});
		let r = P(e), i = P(t);
		return typeof n == "string" && (tn(t, n) || !Ae && is(Ls, n)) ? !1 : Xt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && ns(D, e) && zo(D[e], t);
	}, t.removeHook = function(e, t) {
		if (ns(D, e)) {
			if (t !== void 0) {
				let n = Lo(D[e], t);
				return n === -1 ? void 0 : Vo(D[e], n, 1)[0];
			}
			return Ro(D[e]);
		}
	}, t.removeHooks = function(e) {
		ns(D, e) && (D[e] = []);
	}, t.removeAllHooks = function() {
		D = Vs();
	}, t;
}
var Gs = Ws();
// plugins/chat/format.js
function Ks(e) {
	let t = X.parse(e || "", {
		async: !1,
		gfm: !0,
		breaks: !1
	}), n = Gs.sanitize(t, {
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
function qs(e) {
	return e >= 1e6 ? (e / 1e6).toFixed(1) + "M" : e >= 1e3 ? Math.round(e / 1e3) + "k" : String(e);
}
var Js = (e) => "desk-chat-prompts." + e;
function Ys(e, t) {
	let n = [];
	try {
		let t = JSON.parse(localStorage.getItem(Js(e)) ?? "[]");
		Array.isArray(t) && (n = t.filter((e) => typeof e == "string"));
	} catch {}
	return [...new Set([...t, ...n.filter((e) => !t.includes(e))].filter(Boolean).reverse())].slice(0, 100);
}
function Xs(e, t) {
	try {
		let n = Ys(e, []).filter((e) => e !== t);
		localStorage.setItem(Js(e), JSON.stringify([...n.reverse(), t].slice(-100)));
	} catch {}
}
// plugins/chat/Block.svelte
var Zs = /* @__PURE__ */ G("<details class=\"prompt svelte-5n3hzh\"><summary class=\"svelte-5n3hzh\"> </summary><pre class=\"svelte-5n3hzh\"> </pre></details>"), Qs = /* @__PURE__ */ G("<div class=\"md svelte-5n3hzh\"></div>"), $s = /* @__PURE__ */ G("<div class=\"user svelte-5n3hzh\" data-testid=\"block-user\"><span class=\"tag svelte-5n3hzh\"> </span> <!></div>"), ec = /* @__PURE__ */ G("<div data-testid=\"block-assistant\"></div>"), tc = /* @__PURE__ */ G("<details class=\"reasoning svelte-5n3hzh\" data-testid=\"block-reasoning\"><summary class=\"svelte-5n3hzh\">Thinking</summary> <pre class=\"svelte-5n3hzh\"> </pre></details>"), nc = /* @__PURE__ */ G("<pre class=\"io svelte-5n3hzh\"> </pre>"), rc = /* @__PURE__ */ G("<pre class=\"io out svelte-5n3hzh\"> </pre>"), ic = /* @__PURE__ */ G("<!> <!>", 1), ac = /* @__PURE__ */ G("<div class=\"tool svelte-5n3hzh\" data-testid=\"block-tool\"><button class=\"tool-head svelte-5n3hzh\"><span class=\"tname svelte-5n3hzh\"> </span> <span class=\"tintent svelte-5n3hzh\"> </span> <span class=\"tstate svelte-5n3hzh\" data-testid=\"tool-state\"> </span></button> <!></div>"), oc = /* @__PURE__ */ G("<p class=\"error svelte-5n3hzh\" data-testid=\"block-error\"> </p>"), sc = /* @__PURE__ */ G("<pre class=\"svelte-5n3hzh\"> </pre>"), cc = /* @__PURE__ */ G("<details class=\"system svelte-5n3hzh\" data-testid=\"block-system\"><summary class=\"svelte-5n3hzh\"> </summary> <!></details>");
function lc(e, t) {
	Ue(t, !0);
	let n = /* @__PURE__ */ I(!1);
	function r(e, t = 140) {
		let n = (e || "").replace(/\s+/g, " ").trim();
		return n.length > t ? n.slice(0, t) + "…" : n;
	}
	function i(e) {
		try {
			return JSON.stringify(JSON.parse(e), null, 2);
		} catch {
			return e;
		}
	}
	function a(e) {
		let t = /"intent"\s*:\s*"((?:[^"\\]|\\.)*)"/.exec(e || "");
		return t ? t[1] : "";
	}
	let o = {
		operator: "operator",
		desk: "Desk user",
		controller: "controller prompt",
		message: "message"
	};
	var s = Cr(), c = nn(s), l = (e) => {
		var n = $s(), i = R(n), a = z(i, !0), s = B(i, 2), c = (e) => {
			var n = Zs(), i = R(n), a = z(i, !0), o = z(B(i), !0);
			j(n), V((e) => {
				q(a, e), q(o, t.block.text);
			}, [() => r(t.block.text, 120)]), K(e, n);
		}, l = (e) => {
			var n = Qs();
			Kr(n, () => Ks(t.block.text), !0), j(n), K(e, n);
		};
		J(s, (e) => {
			t.block.source === "controller" ? e(c) : e(l, -1);
		}), j(n), V(() => {
			ei(n, "data-source", t.block.source), q(a, o[t.block.source] || t.block.source);
		}), K(e, n);
	}, u = (e) => {
		var n = ec();
		let r;
		Kr(n, () => Ks(t.block.text), !0), j(n), V(() => r = Yr(n, 1, "assistant md svelte-5n3hzh", null, r, { live: t.block.live })), K(e, n);
	}, d = (e) => {
		var n = tc(), r = z(B(R(n), 2), !0);
		j(n), V(() => {
			n.open = t.block.live, q(r, t.block.text);
		}), K(e, n);
	}, f = (e) => {
		var o = ac(), s = R(o), c = R(s), l = z(c, !0), u = B(c, 2), d = z(u, !0), f = z(B(u, 2), !0);
		j(s);
		var p = B(s, 2), m = (e) => {
			var n = ic(), r = nn(n), a = (e) => {
				var n = nc(), r = z(n, !0);
				V((e) => q(r, e), [() => i(t.block.input)]), K(e, n);
			};
			J(r, (e) => {
				t.block.input && e(a);
			});
			var o = B(r, 2), s = (e) => {
				var n = rc(), r = z(n, !0);
				V(() => q(r, t.block.output)), K(e, n);
			};
			J(o, (e) => {
				t.block.output && e(s);
			}), K(e, n);
		};
		J(p, (e) => {
			W(n) && e(m);
		}), j(o), V((e) => {
			ei(o, "data-state", t.block.state), ei(o, "data-tool", t.block.id), ei(s, "aria-expanded", W(n)), q(l, t.block.name), q(d, e), q(f, t.block.state);
		}, [() => a(t.block.input) || r(t.block.input, 80)]), pr("click", s, () => L(n, !W(n))), K(e, o);
	}, p = (e) => {
		var n = oc(), r = z(n, !0);
		V(() => q(r, t.block.text)), K(e, n);
	}, m = (e) => {
		var n = cc(), i = R(n), a = z(i, !0), o = B(i, 2), s = (e) => {
			var n = sc(), r = z(n, !0);
			V(() => q(r, t.block.text)), K(e, n);
		}, c = /* @__PURE__ */ ft(() => t.block.text.length > 160 || t.block.text.includes("\n"));
		J(o, (e) => {
			W(c) && e(s);
		}), j(n), V((e) => q(a, e), [() => r(t.block.text, 160)]), K(e, n);
	};
	J(c, (e) => {
		t.block.kind === "user" ? e(l) : t.block.kind === "assistant" ? e(u, 1) : t.block.kind === "reasoning" ? e(d, 2) : t.block.kind === "tool" ? e(f, 3) : t.block.kind === "error" ? e(p, 4) : e(m, -1);
	}), K(e, s), We();
}
mr(["click"]);
// kit/attention.js
function uc(e, t, n) {
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
async function dc(e, t) {
	if (!await e.confirm(t.ask)) return !1;
	try {
		return await t.run(), e.toast(t.label + ": done"), !0;
	} catch (n) {
		return e.toast(t.label + " failed: " + String(n.message || n)), !1;
	}
}
// plugins/chat/Pane.svelte
var fc = /* @__PURE__ */ G("<span> </span>"), pc = /* @__PURE__ */ G("· <a class=\"mono svelte-wy7o3a\"> </a>", 1), mc = /* @__PURE__ */ G("<button class=\"older svelte-wy7o3a\" data-testid=\"load-older\"> </button>"), hc = /* @__PURE__ */ G("<p class=\"err svelte-wy7o3a\"> </p>"), gc = /* @__PURE__ */ G("<p class=\"none svelte-wy7o3a\">No messages yet.</p>"), _c = /* @__PURE__ */ G("<div class=\"pending svelte-wy7o3a\" data-testid=\"chat-pending\"><div class=\"ptext svelte-wy7o3a\"> </div><span class=\"pstate svelte-wy7o3a\"> </span></div>"), vc = /* @__PURE__ */ G("<p class=\"status svelte-wy7o3a\" data-testid=\"chat-working\"><span class=\"spin svelte-wy7o3a\"></span> </p>"), yc = /* @__PURE__ */ G("<label class=\"toggle svelte-wy7o3a\" data-testid=\"answer-toggle\"><input type=\"checkbox\"/> answer the question</label>"), bc = /* @__PURE__ */ G("<button type=\"button\" data-testid=\"resolve\" class=\"svelte-wy7o3a\"> </button>"), xc = /* @__PURE__ */ G("<div class=\"item svelte-wy7o3a\" data-testid=\"chat-item\"><div class=\"itxt svelte-wy7o3a\"><span class=\"pill svelte-wy7o3a\"> </span> <span class=\"ititle\"> </span></div> <div class=\"iacts svelte-wy7o3a\"><!> <!></div></div>"), Sc = /* @__PURE__ */ G("<span class=\"model svelte-wy7o3a\"> </span>"), Cc = /* @__PURE__ */ G("<span class=\"usage svelte-wy7o3a\"> </span>"), wc = /* @__PURE__ */ G("<span class=\"notice svelte-wy7o3a\" data-testid=\"chat-notice\"> </span>"), Tc = /* @__PURE__ */ G("<button type=\"button\" class=\"interrupt svelte-wy7o3a\" title=\"Cancel the turn and send now (Ctrl+Enter)\" data-testid=\"chat-interrupt\">Interrupt</button>"), Ec = /* @__PURE__ */ G("<button type=\"button\" class=\"stop svelte-wy7o3a\" title=\"Cancel the turn (Esc)\" data-testid=\"chat-stop\">Stop</button> <!>", 1), Dc = /* @__PURE__ */ G("<form class=\"composer svelte-wy7o3a\"><textarea rows=\"2\" aria-label=\"Message\" enterkeyhint=\"send\" data-testid=\"chat-input\" class=\"svelte-wy7o3a\"></textarea> <div class=\"actions svelte-wy7o3a\"><button type=\"button\" class=\"history svelte-wy7o3a\" aria-label=\"Previous sent message\" title=\"Previous sent message (Up arrow)\">↑</button> <button type=\"button\" class=\"history svelte-wy7o3a\" aria-label=\"Next sent message\" title=\"Next sent message (Down arrow)\">↓</button> <span class=\"metadata svelte-wy7o3a\" data-testid=\"chat-usage\"><!> <!></span> <!> <span class=\"grow svelte-wy7o3a\"></span> <!> <button type=\"submit\" class=\"send svelte-wy7o3a\" data-testid=\"chat-send\"> </button></div></form>"), Oc = /* @__PURE__ */ G("<p class=\"ro svelte-wy7o3a\" data-testid=\"chat-readonly\"> </p>"), kc = /* @__PURE__ */ G("<section class=\"pane svelte-wy7o3a\" data-testid=\"chat-pane\"><header class=\"head svelte-wy7o3a\"><button class=\"x close svelte-wy7o3a\" aria-label=\"Close chat\" title=\"Close\" data-testid=\"session-panel-close\">✕</button> <div class=\"who svelte-wy7o3a\"><div class=\"name svelte-wy7o3a\"> </div> <div class=\"meta svelte-wy7o3a\"><span class=\"mono svelte-wy7o3a\"> </span> <!> <!></div></div> <span class=\"state svelte-wy7o3a\" data-testid=\"chat-status\"> </span></header> <div class=\"scroll svelte-wy7o3a\" data-testid=\"journal\"><!> <!> <!> <!> <!></div> <!> <!></section>");
function Ac(e, t) {
	Ue(t, !0);
	let n = fi(t, "context", 19, () => ({})), r = /* @__PURE__ */ ft(() => String(n().session || "")), i = /* @__PURE__ */ ft(() => String(n().item || "")), a = /* @__PURE__ */ ft(() => W(i).split(":")[0] || ""), o = /* @__PURE__ */ ft(() => W(i).slice(W(a).length + 1)), s = /* @__PURE__ */ I(null);
	hn(() => {
		if (!W(r)) return;
		let e = new _i(t.api, W(r));
		return L(s, e, !0), e.open(), () => e.close();
	});
	let c = /* @__PURE__ */ I(null), l = /* @__PURE__ */ I(""), u = /* @__PURE__ */ I(qt({
		write: !1,
		reason: ""
	}));
	hn(() => {
		if (!W(r)) return;
		let e = !1;
		return L(c, null), L(l, ""), L(u, {
			write: !1,
			reason: ""
		}, !0), t.api.get("api/sessions?since=720h").then((t) => {
			e || L(c, t.find((e) => e.id === W(r)) || null, !0);
		}).catch(() => {}), t.api.get("api/chat").then((t) => {
			if (e) return;
			let n = (t.sessions || []).find((e) => e.id === W(r));
			L(l, n?.alive ? n.key : "", !0);
		}).catch(() => {}), t.api.session.access(W(r)).then((t) => {
			e || L(u, {
				write: !!t.write,
				reason: t.reason || ""
			}, !0);
		}).catch(() => {}), () => {
			e = !0;
		};
	});
	let d = /* @__PURE__ */ I(null);
	hn(() => {
		if (L(d, null), !W(o) || [
			"pool",
			"deploy",
			"rollback",
			"config",
			"config-repo",
			"ingest",
			"resource"
		].includes(W(a))) return;
		let e = !1;
		return t.api.get("api/ticket/" + encodeURIComponent(W(o))).then((t) => {
			e || L(d, t.ticket, !0);
		}).catch(() => {}), () => {
			e = !0;
		};
	});
	let f = /* @__PURE__ */ ft(() => !!W(s)?.busy), p = /* @__PURE__ */ ft(() => {
		let e = W(s)?.blocks ?? [];
		for (let t = e.length - 1; t >= 0; t--) {
			let n = e[t];
			if (n.kind === "tool" && (n.state === "running" || n.state === "input")) return n.name;
		}
		return "";
	}), m = /* @__PURE__ */ ft(() => W(u).write || !!W(l)), h = /* @__PURE__ */ ft(() => W(l).startsWith("escalation:") ? "Escalation " + W(l).slice(11) : W(l).startsWith("attention:") ? "Resolve " + W(i) : W(l).startsWith("desk:") ? "Desk chat" : W(c)?.ticket?.title || "Session"), g = /* @__PURE__ */ I(!0), _ = /* @__PURE__ */ ft(() => W(a) === "question" && W(g)), v = /* @__PURE__ */ I(""), y = /* @__PURE__ */ I(!1), b = /* @__PURE__ */ I(""), x = /* @__PURE__ */ I(qt([]));
	hn(() => {
		let e = (W(s)?.blocks ?? []).filter((e) => e.kind === "user").map((e) => e.text || "");
		W(x).length && W(x).some((t) => e.some((e) => e.includes(t.text))) && L(x, W(x).filter((t) => !e.some((e) => e.includes(t.text))), !0);
	});
	let S = /* @__PURE__ */ I(void 0), C = /* @__PURE__ */ I(!0), w = /* @__PURE__ */ I(void 0), ee = /* @__PURE__ */ I(-1), te = [], T = "", E = null;
	function ne(e) {
		if (W(ee) < 0) {
			if (!e) return;
			te = Ys(W(r), W(s)?.prompts() ?? []), T = W(v);
		}
		let t = e ? Math.min(W(ee) + 1, te.length - 1) : W(ee) - 1;
		t < -1 || !te.length || (L(ee, t, !0), L(v, t < 0 ? T : te[t], !0), ir().then(() => {
			W(w)?.focus(), W(w)?.setSelectionRange(W(v).length, W(v).length), ae();
		}));
	}
	async function re() {
		W(s) && W(S) && (L(C, !1), E = {
			height: W(S).scrollHeight,
			top: W(S).scrollTop
		}, await W(s).loadOlder(), await ir(), E && W(S) && (W(S).scrollTop = E.top + W(S).scrollHeight - E.height, E = null));
	}
	function ie() {
		W(S) && L(C, W(S).scrollHeight - W(S).scrollTop - W(S).clientHeight < 80);
	}
	hn(() => {
		W(s)?.version, W(f), W(C) && !E && ir().then(() => {
			W(S) && (W(S).scrollTop = W(S).scrollHeight);
		});
	});
	function ae() {
		W(w) && (W(w).style.height = "auto", W(w).style.height = Math.min(W(w).scrollHeight, Math.round(window.innerHeight * .4)) + "px");
	}
	async function oe(e) {
		let n = W(v).trim();
		if (n && !W(y)) {
			L(y, !0), L(b, "");
			try {
				W(_) ? (await t.api.call("question.answer", {
					id: W(o),
					text: n
				}), L(b, "Answer sent to the asking session")) : (await t.api.session.send(W(l) ? {
					key: W(l),
					text: n,
					interrupt: e
				} : {
					session: W(r),
					text: n,
					interrupt: e
				}), L(x, [...W(x), {
					text: n,
					state: W(f) && !e ? "queued for after this turn" : "sent, waiting for the session"
				}], !0), W(f) && !e && L(b, "Queued for after this turn")), Xs(W(r), n), L(ee, -1), L(v, ""), L(C, !0), await ir(), ae();
			} catch (e) {
				L(b, String(e.message || e), !0);
			} finally {
				L(y, !1);
			}
		}
	}
	async function se() {
		L(b, "");
		try {
			(await t.api.session.cancel(W(r))).cancelled || L(b, "No turn running");
		} catch (e) {
			L(b, String(e.message || e), !0);
		}
	}
	function ce(e) {
		if (e.isComposing) return;
		let t = e.currentTarget;
		e.key === "ArrowUp" && !e.shiftKey && t.selectionStart === t.selectionEnd && (W(ee) >= 0 || !W(v).slice(0, t.selectionStart).includes("\n")) ? (e.preventDefault(), ne(!0)) : e.key === "ArrowDown" && !e.shiftKey && W(ee) >= 0 ? (e.preventDefault(), ne(!1)) : e.key === "Enter" && !e.shiftKey ? (e.preventDefault(), oe(e.ctrlKey || e.metaKey)) : e.key === "Escape" && W(f) && (e.preventDefault(), e.stopPropagation(), se());
	}
	let le = /* @__PURE__ */ ft(() => uc(t.api, W(a), W(o))), ue = /* @__PURE__ */ I(!1);
	async function de(e) {
		L(ue, !0);
		try {
			await dc(t.api, e) && t.api.closeSession();
		} finally {
			L(ue, !1);
		}
	}
	var fe = kc(), pe = R(fe), me = R(pe), D = B(me, 2), he = R(D), ge = z(he, !0), _e = B(he, 2), ve = R(_e), ye = z(ve, !0), O = B(ve, 2), be = (e) => {
		var t = fc(), n = z(t);
		V(() => q(n, `· ${W(c).ticket.pool ?? ""}`)), K(e, t);
	};
	J(O, (e) => {
		W(c)?.ticket?.pool && e(be);
	});
	var k = B(O, 2), xe = (e) => {
		var n = pc(), r = B(nn(n)), i = z(r, !0);
		V(() => {
			ei(r, "href", "?ticket=" + W(c).current), q(i, W(c).current);
		}), pr("click", r, (e) => {
			e.preventDefault(), t.api.openTicket(W(c).current);
		}), K(e, n);
	};
	J(k, (e) => {
		W(c)?.current && e(xe);
	}), j(_e), j(D);
	var A = z(B(D, 2), !0);
	j(pe);
	var Se = B(pe, 2), Ce = R(Se), we = (e) => {
		var t = mc(), n = z(t, !0);
		V(() => {
			t.disabled = W(s).loadingOlder, q(n, W(s).loadingOlder ? "Loading history…" : `Load older messages (${W(s).earlier} earlier)`);
		}), pr("click", t, re), K(e, t);
	};
	J(Ce, (e) => {
		W(s) && W(s).earlier > 0 && e(we);
	});
	var Te = B(Ce, 2), Ee = (e) => {
		var t = hc(), n = z(t, !0);
		V(() => q(n, W(s).error)), K(e, t);
	};
	J(Te, (e) => {
		W(s)?.error && e(Ee);
	});
	var De = B(Te, 2);
	Br(De, 17, () => W(s)?.blocks ?? [], Ir, (e, t) => {
		lc(e, { get block() {
			return W(t);
		} });
	}, (e) => {
		K(e, gc());
	});
	var Oe = B(De, 2);
	Br(Oe, 17, () => W(x), Ir, (e, t) => {
		var n = _c(), r = R(n), i = z(r, !0), a = z(B(r), !0);
		j(n), V(() => {
			q(i, W(t).text), q(a, W(t).state);
		}), K(e, n);
	});
	var ke = B(Oe, 2), Ae = (e) => {
		var t = vc(), n = B(R(t), 1, !0);
		j(t), V(() => q(n, W(p) || "Working")), K(e, t);
	};
	J(ke, (e) => {
		W(f) && e(Ae);
	}), j(Se), li(Se, (e) => L(S, e), () => W(S));
	var je = B(Se, 2), Me = (e) => {
		var t = xc(), n = R(t), r = R(n), i = z(r, !0), s = z(B(r, 2), !0);
		j(n);
		var c = B(n, 2), l = R(c), u = (e) => {
			var t = yc(), n = R(t);
			$r(n), M(), j(t), ai(n, () => W(g), (e) => L(g, e)), K(e, t);
		};
		J(l, (e) => {
			W(a) === "question" && e(u);
		}), Br(B(l, 2), 17, () => W(le), (e) => e.label, (e, t) => {
			var n = bc(), r = z(n, !0);
			V(() => {
				n.disabled = W(ue), q(r, W(t).label);
			}), pr("click", n, () => de(W(t))), K(e, n);
		}), j(c), j(t), V(() => {
			q(i, W(a)), q(s, W(d)?.title || W(o));
		}), K(e, t);
	};
	J(je, (e) => {
		W(i) && e(Me);
	});
	var Ne = B(je, 2), Pe = (e) => {
		var t = Dc(), n = R(t);
		et(n), li(n, (e) => L(w, e), () => W(w));
		var r = B(n, 2), i = R(r), a = B(i, 2), o = B(a, 2), l = R(o), u = (e) => {
			var t = Sc(), n = z(t, !0);
			V(() => q(n, W(c).model)), K(e, t);
		};
		J(l, (e) => {
			W(c)?.model && e(u);
		});
		var d = B(l, 2), p = (e) => {
			var t = Cc(), n = z(t);
			V((e, t) => q(n, `${e ?? ""} / ${t ?? ""} tokens`), [() => qs(W(s).usage.used), () => qs(W(s).usage.size)]), K(e, t);
		};
		J(d, (e) => {
			W(s)?.usage && e(p);
		}), j(o);
		var m = B(o, 2), h = (e) => {
			var t = wc(), n = z(t, !0);
			V(() => q(n, W(b))), K(e, t);
		};
		J(m, (e) => {
			W(b) && e(h);
		});
		var g = B(m, 4), x = (e) => {
			var t = Ec(), n = nn(t), r = B(n, 2), i = (e) => {
				var t = Tc();
				V((e) => t.disabled = e, [() => !W(v).trim() || W(y)]), pr("click", t, () => oe(!0)), K(e, t);
			};
			J(r, (e) => {
				W(_) || e(i);
			}), pr("click", n, se), K(e, t);
		};
		J(g, (e) => {
			W(f) && e(x);
		});
		var S = B(g, 2), C = z(S, !0);
		j(r), j(t), V((e) => {
			ei(n, "placeholder", W(_) ? "Answer the question" : W(f) ? "Queued for after this turn" : "Message"), a.disabled = W(ee) < 0, S.disabled = e, q(C, W(_) ? "Answer" : W(f) ? "Queue" : "Send");
		}, [() => !W(v).trim() || W(y)]), fr("submit", t, (e) => {
			e.preventDefault(), oe(!1);
		}), pr("input", n, () => {
			L(ee, -1), ae();
		}), pr("keydown", n, ce), ii(n, () => W(v), (e) => L(v, e)), pr("click", i, () => ne(!0)), pr("click", a, () => ne(!1)), K(e, t);
	}, Fe = (e) => {
		var t = Oc(), n = z(t);
		V(() => q(n, `Read-only: ${(W(u).reason || "this session does not accept input") ?? ""}`)), K(e, t);
	};
	J(Ne, (e) => {
		W(m) || W(_) ? e(Pe) : e(Fe, -1);
	}), j(fe), V(() => {
		ei(fe, "aria-label", W(h)), q(ge, W(h)), q(ye, W(r)), q(A, W(s)?.status ?? "connecting");
	}), pr("click", me, () => t.api.closeSession()), fr("scroll", Se, ie), K(e, fe), We();
}
mr([
	"click",
	"input",
	"keydown"
]);
// plugins/chat/ChatList.svelte
var jc = /* @__PURE__ */ G("<div class=\"err svelte-162crgi\"> </div>"), Mc = /* @__PURE__ */ G("<a><span></span> <span class=\"name svelte-162crgi\"> </span> <span class=\"mono faint id svelte-162crgi\"> </span></a>"), Nc = /* @__PURE__ */ G("<div class=\"faint pad0 svelte-162crgi\">No open operator sessions.</div>"), Pc = /* @__PURE__ */ G("<div class=\"card empty\" data-testid=\"escalations-empty\"><div class=\"muted\">No escalation sessions.</div></div>"), Fc = /* @__PURE__ */ G("<tr data-testid=\"escalation-row\"><td class=\"c-title\"> <div class=\"faint mono\"> <!></div></td><td class=\"c-st\"><span> </span></td><td class=\"muted full\"> </td><td class=\"c-st\"><span> </span></td><td class=\"muted c-time\"> </td></tr>"), Ic = /* @__PURE__ */ G("<div class=\"card scroll\"><table class=\"stack\" data-testid=\"escalations\"><thead><tr><th>Ticket</th><th>Cause</th><th>Why</th><th>State</th><th>Age</th></tr></thead><tbody></tbody></table></div>"), Lc = /* @__PURE__ */ G("<header class=\"vh svelte-162crgi\"><h1>Chat</h1> <span class=\"sub svelte-162crgi\">pool <span class=\"mono\"> </span></span></header> <div class=\"list svelte-162crgi\"><form class=\"new svelte-162crgi\"><textarea rows=\"3\" placeholder=\"What should the operator look at?\" data-testid=\"chat-new-input\" class=\"svelte-162crgi\"></textarea> <div class=\"row svelte-162crgi\"><button type=\"submit\" class=\"primary\" data-testid=\"chat-new\">New session</button></div></form> <!> <div data-testid=\"chat-list\"></div> <h2 class=\"st\">Escalations</h2> <!></div>", 1);
function Rc(e, t) {
	Ue(t, !0);
	let n = /* @__PURE__ */ I(null), r = /* @__PURE__ */ I(null), i = /* @__PURE__ */ I(""), a = /* @__PURE__ */ I(""), o = /* @__PURE__ */ I(!1), s = /* @__PURE__ */ I("");
	async function c() {
		try {
			await (async (e) => {
				var t = y(e, 2);
				L(n, t[0], !0), L(r, t[1], !0);
			})(await Promise.all([t.api.get("api/chat"), t.api.get("api/escalations")])), L(i, "");
		} catch (e) {
			L(i, String(e.message || e), !0);
		}
	}
	hn(() => {
		L(s, t.api.params().session || "", !0), c();
		let e = 0, n = t.api.subscribe(() => {
			L(s, t.api.params().session || "", !0), e ||= window.setTimeout(() => {
				e = 0, c();
			}, 1e3);
		});
		return () => {
			n(), clearTimeout(e);
		};
	});
	let l = /* @__PURE__ */ ft(() => (W(n)?.sessions ?? []).filter((e) => !e.key.startsWith("escalation:"))), u = /* @__PURE__ */ ft(() => (W(r) ?? []).filter((e) => e.open)), d = /* @__PURE__ */ ft(() => (W(r) ?? []).filter((e) => !e.open));
	function f(e) {
		return e.startsWith("attention:") ? "Resolve " + e.slice(10) : e.startsWith("desk:") ? "Desk chat" : e.split(":")[0] || "Session";
	}
	let p = {
		run: "run failed",
		pool: "pool quarantined",
		wait: "wait timed out",
		guard: "needs approval",
		label: "needs human"
	};
	function m(e) {
		if (!e || e.startsWith("0001")) return "";
		let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3);
		return t < 60 ? t + "s ago" : t < 3600 ? Math.round(t / 60) + "m ago" : t < 86400 ? Math.round(t / 3600) + "h ago" : Math.round(t / 86400) + "d ago";
	}
	function h(e, n) {
		L(s, e, !0), t.api.openSession(e, n ? { item: n } : void 0);
	}
	async function g() {
		if (W(a).trim()) {
			L(o, !0), L(i, "");
			try {
				let e = await t.api.post("api/chat/new", { text: W(a) });
				L(a, ""), h(e.session);
			} catch (e) {
				L(i, String(e.message || e), !0);
			} finally {
				L(o, !1);
			}
		}
	}
	var _ = Lc(), v = nn(_), b = B(R(v), 2), x = z(B(R(b)), !0);
	j(b), j(v);
	var S = B(v, 2), C = R(S), w = R(C);
	et(w);
	var ee = z(B(w, 2));
	j(C);
	var te = B(C, 2), T = (e) => {
		var t = jc(), n = z(t, !0);
		V(() => q(n, W(i))), K(e, t);
	};
	J(te, (e) => {
		W(i) && e(T);
	});
	var E = B(te, 2);
	Br(E, 21, () => W(l), (e) => e.id, (e, t) => {
		var n = Mc();
		let r;
		var i = R(n);
		let a;
		var o = B(i, 2), c = z(o, !0), l = z(B(o, 2), !0);
		j(n), V((e) => {
			ei(n, "href", "?view=chat&session=" + W(t).id), r = Yr(n, 1, "item svelte-162crgi", null, r, { active: W(s) === W(t).id }), a = Yr(i, 1, "dot", null, a, { accent: W(t).alive }), q(c, e), q(l, W(t).id);
		}, [() => f(W(t).key)]), pr("click", n, (e) => {
			e.preventDefault(), h(W(t).id, W(t).key.startsWith("attention:") ? W(t).key.slice(10) : "");
		}), K(e, n);
	}, (e) => {
		K(e, Nc());
	}), j(E);
	var ne = B(E, 4), re = (e) => {
		K(e, Pc());
	}, ie = (e) => {
		var t = Ic(), n = R(t), r = B(R(n));
		Br(r, 21, () => [...W(u), ...W(d)], (e) => e.ticket, (e, t) => {
			var n = Fc();
			let r;
			var i = R(n), a = R(i, !0), o = B(a), c = R(o, !0), l = B(c), u = (e) => {
				var n = Sr();
				V(() => q(n, `· ${W(t).held ?? ""}`)), K(e, n);
			};
			J(l, (e) => {
				W(t).held && W(t).held !== W(t).ticket && e(u);
			}), j(o), j(i);
			var d = B(i), f = R(d), g = z(f, !0);
			j(d);
			var _ = B(d), v = z(_, !0), y = B(_), b = R(y), x = z(b, !0);
			j(y);
			var S = z(B(y), !0);
			j(n), V((e) => {
				r = Yr(n, 1, "click", null, r, { sel: W(s) === W(t).session }), q(a, W(t).title), q(c, W(t).ticket), Yr(f, 1, `pill ${W(t).open ? "warn" : ""}`), q(g, p[W(t).cause] ?? W(t).cause ?? "escalation"), q(v, W(t).why), Yr(b, 1, `pill ${W(t).open ? "accent" : "ok"}`), q(x, W(t).open ? "open" : "closed" + (W(t).resolution ? " · " + W(t).resolution : "")), q(S, e);
			}, [() => m(W(t).open ? W(t).since : W(t).closed)]), pr("click", n, () => h(W(t).session, "ticket:" + W(t).ticket)), K(e, n);
		}), j(r), j(n), j(t), K(e, t);
	};
	J(ne, (e) => {
		W(r) && W(u).length + W(d).length === 0 ? e(re) : W(r) && e(ie, 1);
	}), j(S), V((e) => {
		q(x, W(n)?.pool ?? "…"), ee.disabled = e;
	}, [() => W(o) || !W(a).trim()]), fr("submit", C, (e) => {
		e.preventDefault(), g();
	}), ii(w, () => W(a), (e) => L(a, e)), K(e, _), We();
}
mr(["click"]);
// plugins/chat/DrawerChat.svelte
var zc = /* @__PURE__ */ G("<div class=\"wrap svelte-1i9xi0d\" data-testid=\"escalation-chat\"><div class=\"bar svelte-1i9xi0d\"><a class=\"svelte-1i9xi0d\">open in side panel</a></div> <div class=\"frame svelte-1i9xi0d\"><!></div></div>");
function Bc(e, t) {
	Ue(t, !0);
	let n = fi(t, "context", 19, () => ({})), r = /* @__PURE__ */ ft(() => n().ticket || {}), i = (e) => typeof e?.meta == "string" ? JSON.parse(e.meta) : e?.meta || {}, a = /* @__PURE__ */ ft(() => {
		if (W(r).kind === "session") return W(r).id;
		let e = i(W(r)).escalation || i(W(r)).gateway;
		return e?.session ? e.session : W(r).holder && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(W(r).holder) ? W(r).holder : "";
	});
	var o = Cr(), s = nn(o), c = (e) => {
		var n = zc(), r = R(n), i = z(r), o = B(r, 2), s = R(o);
		{
			let e = /* @__PURE__ */ ft(() => ({ session: W(a) }));
			Ac(s, {
				get api() {
					return t.api;
				},
				get context() {
					return W(e);
				}
			});
		}
		j(o), j(n), V(() => ei(i, "href", "?session=" + W(a))), pr("click", i, (e) => {
			e.preventDefault(), t.api.openSession(W(a));
		}), K(e, n);
	};
	J(s, (e) => {
		W(a) && e(c);
	}), K(e, o), We();
}
mr(["click"]);
// plugins/chat/index.svelte.js
var Vc = new URL("./style.css", import.meta.url).href;
function Hc(e) {
	return class extends HTMLElement {
		#e = /* @__PURE__ */ I(qt({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return W(this.#e);
		}
		set props(e) {
			L(this.#e, e, !0);
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
			if (!document.querySelector(`link[href="${Vc}"]`)) {
				let e = document.createElement("link");
				e.rel = "stylesheet", e.href = Vc, document.head.append(e);
			}
			this.style.display = "contents", this.app = Ar(e, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && Pr(this.app), this.app = null;
		}
	};
}
for (let [e, t] of [
	["desk-chat-pane", Ac],
	["desk-chat-view", Rc],
	["desk-chat-drawer", Bc]
]) customElements.get(e) || customElements.define(e, Hc(t));
