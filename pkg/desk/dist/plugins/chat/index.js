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
var b = 1024, x = 2048, S = 4096, C = 8192, w = 16384, ee = 32768, te = 1 << 25, T = 65536, ne = 1 << 19, re = 1 << 20, ie = 1 << 25, ae = 1 << 21, oe = 1 << 22, se = 1 << 23, ce = Symbol("$state"), le = Symbol("component"), ue = Symbol("legacy props"), de = Symbol(""), fe = Symbol("attributes"), pe = Symbol("class"), me = Symbol("style"), he = Symbol("text"), E = Symbol("form reset"), ge = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), _e = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function ve() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function ye(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function be() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var D = !1;
function xe(e) {
	D = e;
}
var O;
function Se(t) {
	if (t === null) throw ye(), e;
	return O = t;
}
function k() {
	return Se(/* @__PURE__ */ tn(O));
}
function A(t) {
	if (D) {
		if (/* @__PURE__ */ tn(O) !== null) throw ye(), e;
		O = t;
	}
}
function j(e = 1) {
	if (D) {
		for (var t = e, n = O; t--;) n = /* @__PURE__ */ tn(n);
		O = n;
	}
}
function Ce(e = !0) {
	for (var t = 0, n = O;;) {
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
function we(t) {
	if (!t || t.nodeType !== 8) throw ye(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function Te(e) {
	return e === this.v;
}
function Ee(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function De(e) {
	return !Ee(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function Oe() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function ke(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function Ae(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function je() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function Me(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Ne() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Pe(e) {
	throw Error("https://svelte.dev/e/props_invalid_value");
}
function Fe() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Ie() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Le() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Re() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/shared/clone.js
var ze = [];
function Be(e, t = !1, n = !1) {
	return Ve(e, /* @__PURE__ */ new Map(), "", ze, null, n);
}
function Ve(e, t, n, r, i = null, o = !1) {
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
				l in e && (c[l] = Ve(u, t, n, r, null, o));
			}
			return c;
		}
		if (m(e) === f) {
			c = {}, t.set(e, c), i !== null && t.set(i, c);
			for (var d of Object.keys(e)) c[d] = Ve(e[d], t, n, r, null, o);
			return c;
		}
		if (e instanceof Date) return e.getTime(), structuredClone(e);
		if (typeof e.toJSON == "function" && !o) return Ve(e.toJSON(), t, n, r, e);
	}
	if (e instanceof EventTarget) return e;
	try {
		return structuredClone(e);
	} catch {
		return e;
	}
}
// node_modules/svelte/src/internal/client/context.js
var He = null;
function Ue(e) {
	He = e;
}
function We(e, t = !1, n) {
	He = {
		p: He,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: U,
		l: null
	};
}
function Ge(e) {
	var t = He, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) gn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, He = t.p, Ke(e);
}
function Ke(e = {}) {
	return l(e, le, { value: !0 }), e;
}
function qe() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var Je = [];
function Ye() {
	var e = Je;
	Je = [], _(e);
}
function Xe(e) {
	if (Je.length === 0 && !St) {
		var t = Je;
		queueMicrotask(() => {
			t === Je && Ye();
		});
	}
	Je.push(e);
}
function Ze() {
	for (; Je.length > 0;) Ye();
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Qe = ~(x | S | b);
function M(e, t) {
	e.f = e.f & Qe | t;
}
function $e(e) {
	e.f & 512 || e.deps === null ? M(e, b) : M(e, S);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function et(e, t) {
	e.f & 2048 ? t.set(e, x) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, S), M(e, b);
}
// node_modules/svelte/src/internal/client/dom/elements/misc.js
function tt(e) {
	D && /* @__PURE__ */ en(e) !== null && rn(e);
}
var nt = !1;
function rt() {
	nt || (nt = !0, document.addEventListener("reset", (e) => {
		Promise.resolve().then(() => {
			if (!e.defaultPrevented) for (let t of e.target.elements) t[E]?.();
		});
	}, { capture: !0 }));
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function it(e) {
	var t = H, n = U;
	Rn(null), zn(null);
	try {
		return e();
	} finally {
		Rn(t), zn(n);
	}
}
function at(e, t, n, r = n) {
	e.addEventListener(t, () => it(n));
	let i = e[E];
	e[E] = i ? () => {
		i(), r(!0);
	} : () => r(!0), rt();
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function ot(e, t, n, r) {
	let i = qe() ? ut : P;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = U, c = st(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				ln(e, s);
			}
			ct();
		}
	}
	var d = lt();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ ft(e))).then(u).catch((e) => ln(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), ct());
	}) : f();
}
function st() {
	var e = U, t = H, n = He, r = F;
	return function(i = !0) {
		zn(e), Rn(t), Ue(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function ct(e = !0) {
	zn(null), Rn(null), Ue(null), e && F?.deactivate();
}
function lt() {
	var e = U, t = e.b, n = F, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function ut(e) {
	var n = 2 | x;
	return U !== null && (U.f |= ne), {
		ctx: He,
		deps: null,
		effects: null,
		equals: Te,
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
var dt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function ft(e, n, r) {
	let i = U;
	i === null && Oe();
	var a = void 0, o = zt(t), s = !H, c = /* @__PURE__ */ new Set();
	return yn(() => {
		var t = U, n = v();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== ge && n.reject(e);
			}).finally(ct);
		} catch (e) {
			n.reject(e), ct();
		}
		var r = F;
		if (s) {
			if (t.f & 32768) var l = lt();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(dt);
			else for (let e of c.values()) e.reject(dt);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== dt && (r.activate(), t ? (o.f |= se, Ut(o, t)) : (o.f & 8388608 && (o.f ^= se), Ut(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), mn(() => {
		for (let e of c) e.reject(dt);
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
function N(e) {
	let t = /* @__PURE__ */ ut(e);
	return Vn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function P(e) {
	let t = /* @__PURE__ */ ut(e);
	return t.equals = De, t;
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
	if (!Fn && i !== null && e.v !== t && i.f & 24576) return ve(), e.v;
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
	!e.equals(t) && (e.wv = Xn(), (!F?.is_fork || e.deps === null) && (F === null ? e.v = t : (F.capture(e, t, !0), yt?.capture(e, t, !0)), e.deps === null)) ? M(e, b) : Fn || (bt === null ? $e(e) : (pn() || F?.is_fork) && bt.set(e, t));
}
function gt(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && it(() => {
		t.ac.abort(ge), t.ac = null;
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
			for (var r of n.d) M(r, x), t(r);
			for (r of n.m) M(r, S), t(r);
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
		for (let [e, t] of this.#s) M(e, t), this.schedule(e);
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), M(i, x), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), F = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) et(e[t], this.#s);
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
		for (let e of this.async_deriveds.values()) e.reject(dt);
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
		this.#u || (this.#u = !0, Xe(() => {
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
			!Ct && !St && Xe(() => {
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
			if (Ze(), F === null) return r;
			F.flush();
		}
	} finally {
		St = t, yt = n;
	}
}
function At() {
	try {
		Ne();
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
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), M(e, b);
		for (var n = e.first; n !== null;) Pt(n, t), n = n.next;
	}
}
function Ft(e) {
	M(e, b);
	for (var t = e.first; t !== null;) Ft(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var It = /* @__PURE__ */ new Set(), Lt = /* @__PURE__ */ new Map(), Rt = !1;
function zt(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: Te,
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
	return t || (r.equals = De), r;
}
function L(e, t, n = !1) {
	return H !== null && (!Ln || H.f & 131072) && qe() && H.f & 4325394 && (Bn === null || !Bn.has(e)) && Le(), Ut(e, n ? qt(t) : t, Tt);
}
var Vt = null, Ht = 0;
function Ut(e, t, n = null) {
	if (!e.equals(t)) {
		Fn ? Lt.set(e, t) : Lt.has(e) || Lt.set(e, e.v);
		var r = Ot.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && mt(t), bt === null && $e(t);
		}
		e.wv = Xn(), Vt = null, Ht = 0, Kt(e, x, n), Vt = null, qe() && U !== null && U.f & 1024 && !(U.f & 96) && (Wn === null ? Gn([e]) : Wn.push(e)), !r.is_fork && It.size > 0 && !Rt && Wt();
	}
	return t;
}
function Wt() {
	Rt = !1;
	for (let e of It) {
		e.f & 1024 && M(e, S);
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
		var i = qe(), a = r.length;
		if (Ht += a, Ht > 1e5 && Vt === null && (Vt = /* @__PURE__ */ new Set()), Vt !== null) {
			if (Vt.has(e)) return;
			Vt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== U) {
				var l = (c & x) === 0;
				if (l && M(s, t), c & 131072) It.add(s);
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
	if (typeof e != "object" || !e || ce in e || le in e) return e;
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
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Fe();
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
			if (i === ce) return e;
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
			if (n === ce) return !0;
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
			Ie();
		}
	});
}
var Jt, Yt, Xt, Zt;
function Qt() {
	if (Jt === void 0) {
		Jt = window, Yt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Xt = u(t, "firstChild").get, Zt = u(t, "nextSibling").get, h(e) && (e[pe] = void 0, e[fe] = null, e[me] = void 0, e.__e = void 0), h(n) && (n[he] = void 0);
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
	if (!D) return /* @__PURE__ */ en(e);
	var n = /* @__PURE__ */ en(O);
	if (n === null) n = O.appendChild($t());
	else if (t && n.nodeType !== 3) {
		var r = $t();
		return n?.before(r), Se(r), r;
	}
	return t && sn(n), Se(n), n;
}
function nn(e, t = !1) {
	if (!D) {
		var n = /* @__PURE__ */ en(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ tn(n) : n;
	}
	if (t) {
		if (O?.nodeType !== 3) {
			var r = $t();
			return O?.before(r), Se(r), r;
		}
		sn(O);
	}
	return O;
}
function z(e, t = !1) {
	if (!D) return /* @__PURE__ */ en(e);
	var n = R(e, t);
	return A(e), n;
}
function B(e, t = 1, n = !1) {
	let r = D ? O : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ tn(r);
	if (!D) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = $t();
			return r === null ? i?.after(a) : r.before(a), Se(a), a;
		}
		sn(r);
	}
	return Se(r), r;
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
	if (t === null) return H.f |= se, e;
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
	U === null && (H === null && Me(e), je()), Fn && Ae(e);
}
function dn(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function fn(e, t) {
	var n = U;
	n !== null && n.f & 8192 && (e |= C);
	var r = {
		ctx: He,
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
	return M(t, b), t.teardown = e, t;
}
function hn(e) {
	un("$effect");
	var t = U.f;
	if (!H && t & 32 && He !== null && !He.i) {
		var n = He;
		(n.e ??= []).push(e);
	} else return gn(e);
}
function gn(e) {
	return fn(4 | re, e);
}
function _n(e) {
	Ot.ensure();
	let t = fn(64 | ne, e);
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
	return fn(oe | ne, e);
}
function bn(e, t = 0) {
	return fn(8 | t, e);
}
function V(e, t = [], n = [], r = []) {
	ot(r, t, n, (t) => {
		fn(8, () => {
			e(...t.map(W));
		});
	});
}
function xn(e, t = 0) {
	return fn(16 | t, e);
}
function Sn(e) {
	return fn(32 | ne, e);
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
		e !== null && it(() => {
			e.abort(ge);
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
		e.f ^= C, e.f & 1024 || (M(e, x), Ot.ensure().schedule(e));
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
		t & 512 && bt === null && M(e, b);
	}
	return !1;
}
function Qn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(Bn !== null && Bn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Qn(a, t, !1) : t === a && (n ? M(a, x) : a.f & 1024 && M(a, S), Nt(a));
	}
}
function $n(e) {
	var t = Hn, n = Un, r = Wn, i = H, a = Bn, o = He, s = Ln, c = Jn, l = e.f;
	Hn = null, Un = 0, Wn = null, H = l & 96 ? null : e, Bn = null, Ue(e.ctx), Ln = !1, Jn = ++qn, e.ac !== null && (it(() => {
		e.ac.abort(ge);
	}), e.ac = null);
	try {
		e.f |= ae;
		var u = e.fn, d = u();
		e.f |= ee;
		var f = er(e);
		if (qe() && Wn !== null && !Ln && f !== null && !(e.f & 6146)) for (var p = 0; p < Wn.length; p++) Qn(Wn[p], e);
		if (i !== null && i !== e) {
			if (qn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = qn;
			if (t !== null) for (let e of t) e.rv = qn;
			Wn !== null && (r === null ? r = Wn : r.push(...Wn));
		}
		return e.f & 8388608 && (e.f ^= se), d;
	} catch (t) {
		return er(e), cn(t);
	} finally {
		e.f ^= ae, Hn = t, Un = n, Wn = r, H = i, Bn = a, Ue(o), Ln = s, Jn = c;
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
		c.f & 512 && (c.f ^= 512), c.v !== t && $e(c), c.ac !== null && it(() => {
			c.ac.abort(ge), c.ac = null, M(c, x);
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
		M(e, b);
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
function cr(e) {
	if (!(typeof e != "object" || !e || e instanceof EventTarget)) {
		if (ce in e) lr(e);
		else if (!Array.isArray(e)) for (let t in e) {
			let n = e[t];
			typeof n == "object" && n && ce in n && lr(n);
		}
	}
}
function lr(e, t = /* @__PURE__ */ new Set()) {
	if (typeof e == "object" && e && !(e instanceof EventTarget) && !t.has(e)) {
		t.add(e), e instanceof Date && e.getTime();
		for (let n in e) try {
			lr(e[n], t);
		} catch {}
		let n = m(e);
		if (n !== Object.prototype && n !== Array.prototype && n !== Map.prototype && n !== Set.prototype && n !== Date.prototype) {
			let t = d(n);
			for (let n in t) {
				let r = t[n].get;
				if (r) try {
					r.call(e);
				} catch {}
			}
		}
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var ur = Symbol("events"), dr = /* @__PURE__ */ new Set(), fr = /* @__PURE__ */ new Set();
function pr(e, t, n, r = {}) {
	function i(e) {
		if (r.capture || yr.call(t, e), !e.cancelBubble) return it(() => n?.call(this, e));
	}
	return e.startsWith("pointer") || e.startsWith("touch") || e === "wheel" ? (i.__removed = !1, Xe(() => {
		i.__removed || t.addEventListener(e, i, r);
	})) : t.addEventListener(e, i, r), i;
}
function mr(e, t, n, r, i) {
	var a = {
		capture: r,
		passive: i
	}, o = pr(e, t, n, a);
	(t === document.body || t === window || t === document || t instanceof HTMLMediaElement) && mn(() => {
		o.__removed = !0, t.removeEventListener(e, o, a);
	});
}
function hr(e, t, n) {
	(t[ur] ??= {})[e] = n;
}
function gr(e) {
	for (var t = 0; t < e.length; t++) dr.add(e[t]);
	for (var n of fr) n(e);
}
var _r = null, vr = !1;
function yr(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	_r = e, vr || (vr = !0, setTimeout(() => {
		vr = !1, _r = null;
	}));
	var o = 0, s = _r === e && e[ur];
	if (s) {
		var c = i.indexOf(s);
		if (c !== -1 && (t === document || t === window)) {
			e[ur] = t;
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
					var h = a[ur]?.[r];
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
			e[ur] = t, delete e.currentTarget, Rn(d), zn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var br = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function xr(e) {
	return br?.createHTML(e) ?? e;
}
function Sr(e) {
	var t = on("template");
	return t.innerHTML = xr(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function Cr(e, t) {
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
		if (D) return Cr(O, null), O;
		i === void 0 && (i = Sr(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ en(i)));
		var t = r || Yt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ en(t), s = t.lastChild;
			Cr(o, s);
		} else Cr(t, t);
		return t;
	};
}
function wr(e = "") {
	if (!D) {
		var t = $t(e + "");
		return Cr(t, t), t;
	}
	var n = O;
	return n.nodeType === 3 ? sn(n) : (n.before(n = $t()), Se(n)), Cr(n, n), n;
}
function Tr() {
	if (D) return Cr(O, null), O;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = $t();
	return e.append(t, n), Cr(t, n), e;
}
function K(e, t) {
	if (D) {
		var n = U;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = O), k();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var Er = ["touchstart", "touchmove"];
function Dr(e) {
	return Er.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function Or(e) {
	let t = 0, n = zt(0), r;
	return () => {
		pn() && (W(n), bn(() => (t === 0 && (r = sr(() => e(() => Gt(n)))), t += 1, () => {
			Xe(() => {
				--t, t === 0 && (r?.(), r = void 0, Gt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var kr = T | ne;
function Ar(e, t, n, r) {
	new jr(e, t, n, r);
}
var jr = class {
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
	#m = Or(() => (this.#p = zt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = U;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = U.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = xn(() => {
			if (D) {
				let e = we(this.#t);
				k();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, kr), D && (this.#e = O);
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
		Xe(r), t && (this.#s = Sn(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? be() : (t = !0, n && Re(), this.#s !== null && kn(this.#s, () => {
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
		e && (this.is_pending = !0, this.#o = Sn(() => e(this.#e)), Xe(() => {
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
		et(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = U, n = H, r = He;
		zn(this.#i), Rn(this.#i), Ue(this.#i.ctx);
		try {
			return Ot.ensure(), e();
		} finally {
			zn(t), Rn(n), Ue(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && kn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Xe(() => {
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
		this.#a &&= (En(this.#a), null), this.#o &&= (En(this.#o), null), this.#s &&= (En(this.#s), null), D && (Se(this.#t), j(), Se(Ce()));
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
		Xe(() => {
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
	n !== (e[he] ??= e.nodeValue) && (e[he] = n, e.nodeValue = `${n}`);
}
function Mr(e, t) {
	return Pr(e, t);
}
var Nr = /* @__PURE__ */ new Map();
function Pr(t, { target: n, anchor: r, props: i = {}, events: a, context: o, intro: s = !0, transformError: l }) {
	Qt();
	var u = void 0, d = _n(() => {
		var s = r ?? n.appendChild($t());
		Ar(s, { pending: () => {} }, (n) => {
			We({});
			var r = He;
			if (o && (r.c = o), a && (i.$$events = a), D && Cr(n, null), u = t(n, i) || Ke(), D && (U.nodes.end = O, O === null || O.nodeType !== 8 || O.data !== "]")) throw ye(), e;
			Ge();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = Dr(r);
					for (let e of [n, document]) {
						var a = Nr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), Nr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, yr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(c(dr)), fr.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = Nr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, yr), t.delete(e), t.size === 0 && Nr.delete(r)) : t.set(e, i);
			}
			fr.delete(f), s !== r && s.parentNode?.removeChild(s);
		};
	});
	return Fr.set(u, d), u;
}
var Fr = /* @__PURE__ */ new WeakMap();
function Ir(e, t) {
	let n = Fr.get(e);
	return n ? (Fr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var Lr = class {
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
		} else D && (this.anchor = O), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function J(e, t, n = !1) {
	var r;
	D && (r = O, k());
	var i = new Lr(e), a = n ? T : 0;
	function o(e, t) {
		if (D) {
			var n = we(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Ce();
				Se(a), i.anchor = a, xe(!1), i.ensure(e, t), xe(!0);
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
function Rr(e, t) {
	return t;
}
function zr(e, t, n) {
	for (var r = [], i = t.length, a, o = t.length, s = 0; s < i; s++) {
		let n = t[s];
		kn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					Br(e, c(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
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
		Br(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function Br(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= ie, Nn(a, document.createDocumentFragment())) : En(t[i], n);
	}
}
var Vr;
function Hr(e, t, n, r, i, o = null) {
	var s = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		s = D ? Se(/* @__PURE__ */ en(u)) : u.appendChild($t());
	}
	D && k();
	var d = null, f = /* @__PURE__ */ P(() => {
		var e = n();
		return a(e) ? e : e == null ? [] : c(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = W(f);
			_.fallback = d, Wr(_, n, s, t, r), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= ie, Kr(d, null, s)) : jn(d) : kn(d, () => {
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
			D && we(s) === "[!" != (a === 0) && (s = Ce(), Se(s), xe(!1), c = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = F, v = an(), y = 0; y < a; y += 1) {
				D && O.nodeType === 8 && O.data === "]" && (s = O, c = !0, xe(!1));
				var b = e[y], x = r(b, y), S = m ? null : l.get(x);
				S ? (S.v && Ut(S.v, b), S.i && Ut(S.i, y), v && _.unskip_effect(S.e)) : (S = Gr(l, m ? s : Vr ??= $t(), b, x, y, i, t, n), m || (S.e.f |= ie), l.set(x, S)), u.add(x);
			}
			if (a === 0 && o && !d && (m ? d = Sn(() => o(s)) : (d = Sn(() => o(Vr ??= $t())), d.f |= ie)), a > u.size && ke("", "", ""), D && a > 0 && Se(Ce()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			c && xe(!0), W(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, D && (s = O);
}
function Ur(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function Wr(e, t, n, r, i) {
	var a = !!(r & 8), o = t.length, s = e.items, l = Ur(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < o; v += 1) h = t[v], g = i(h, v), _ = s.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < o; v += 1) {
		if (h = t[v], g = i(h, v), _ = s.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (jn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= ie, _ === l) Kr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), qr(e, d, _), qr(e, _, y), Kr(_, y, n), d = _, p = [], m = [], l = Ur(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) Kr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					qr(e, S.prev, C.next), qr(e, d, S), qr(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), Kr(_, l, n), qr(e, _.prev, _.next), qr(e, _, d === null ? e.effect.first : d.next), qr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = Ur(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = Ur(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (Br(e, c(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var w = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || w.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && w.push(l), l = Ur(l.next);
		var ee = w.length;
		if (ee > 0) {
			var te = r & 4 && o === 0 ? n : null;
			if (a) {
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.measure();
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.fix();
			}
			zr(e, w, te);
		}
	}
	a && Xe(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Gr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? zt(n) : /* @__PURE__ */ Bt(n, !1, !1) : null, l = o & 2 ? zt(i) : null;
	return {
		v: c,
		i: l,
		e: Sn(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Kr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ tn(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function qr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
function Jr(t, n, a = !1, o = !1, s = !1, c = !1) {
	var l = t, u = "";
	if (a) {
		var d = t;
		D && (l = Se(/* @__PURE__ */ en(d)));
	}
	V(() => {
		var t = U;
		if (u === (u = n() ?? "")) D && k();
		else if (a && !D) t.nodes = null, d.innerHTML = u, u !== "" && Cr(/* @__PURE__ */ en(d), d.lastChild);
		else if (t.nodes !== null && (Dn(t.nodes.start, t.nodes.end), t.nodes = null), u !== "") {
			if (D) {
				for (var c = O.data, f = k(), p = f; f !== null && (f.nodeType !== 8 || f.data !== "");) p = f, f = /* @__PURE__ */ tn(f);
				if (f === null) throw ye(), e;
				Cr(O, p), l = Se(f);
			} else {
				var m = on(o ? "svg" : s ? "math" : "template", o ? r : s ? i : void 0);
				m.innerHTML = u;
				var h = o || s ? m : m.content;
				if (Cr(/* @__PURE__ */ en(h), h.lastChild), o || s) for (; /* @__PURE__ */ en(h);) l.before(/* @__PURE__ */ en(h));
				else l.before(h);
			}
		}
	});
}
// node_modules/svelte/src/internal/client/dom/elements/actions.js
function Yr(e, t, n) {
	vn(() => {
		var r = sr(() => t(e, n?.()) || {});
		if (n && r?.update) {
			var i = !1, a = {};
			bn(() => {
				var e = n();
				cr(e), i && Ee(a, e) && (a = e, r.update(e));
			}), i = !0;
		}
		if (r?.destroy) return () => r.destroy();
	});
}
// node_modules/svelte/src/internal/shared/attributes.js
var Xr = [..." 	\n\r\f\xA0\v﻿"];
function Zr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || Xr.includes(r[o - 1])) && (s === r.length || Xr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Qr(e, t, n, r, i, a) {
	var o = e[pe];
	if (D || o !== n || o === void 0) {
		var s = Zr(n, r, a);
		D && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[pe] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var $r = Symbol("is custom element"), ei = Symbol("is html"), ti = _e ? "link" : "LINK";
function ni(e, t, n, r) {
	var i = ri(e);
	if (!(D && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === ti))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[de] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && ai(e).has(t) ? e[t] = n : (!D || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function ri(e) {
	return e[fe] ??= {
		[$r]: e.nodeName.includes("-"),
		[ei]: e.namespaceURI === n
	};
}
var ii = /* @__PURE__ */ new Map();
function ai(e) {
	var t = e.getAttribute("is") || e.nodeName, n = ii.get(t);
	if (n) return n;
	ii.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = d(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = m(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/input.js
function oi(e, t, n = t) {
	var r = /* @__PURE__ */ new WeakSet();
	at(e, "input", async (i) => {
		var a = i ? e.defaultValue : e.value;
		if (a = si(e) ? ci(a) : a, n(a), F !== null && r.add(F), await ir(), a !== (a = t())) {
			var o = e.selectionStart, s = e.selectionEnd, c = e.value.length;
			if (e.value = a ?? "", s !== null) {
				var l = e.value.length;
				o === s && s === c && l > c ? (e.selectionStart = l, e.selectionEnd = l) : (e.selectionStart = o, e.selectionEnd = Math.min(s, l));
			}
		}
	}), (D && e.defaultValue !== e.value || sr(t) == null && e.value) && (n(si(e) ? ci(e.value) : e.value), F !== null && r.add(F)), bn(() => {
		var n = t();
		if (e === document.activeElement) {
			var i = F;
			if (r.has(i)) return;
		}
		si(e) && n === ci(e.value) || (e.type !== "date" || n || e.value) && n !== e.value && (e.value = n ?? "");
	});
}
function si(e) {
	var t = e.type;
	return t === "number" || t === "range";
}
function ci(e) {
	return e === "" ? null : +e;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/this.js
function li(e, t) {
	return e === t || e?.[ce] === t;
}
function ui(e = Ke(), t, n, r) {
	var i = He.r, a = U;
	return vn(() => {
		var o, s;
		return bn(() => {
			o = s, s = r?.() || [], sr(() => {
				li(n(...s), e) || (t(e, ...s), o && li(n(...o), e) && t(null, ...o));
			});
		}), () => {
			let r = a;
			for (; r !== i && r.parent !== null && r.parent.f & 33554432;) r = r.parent;
			let o = () => {
				s && li(n(...s), e) && t(null, ...s);
			}, c = r.teardown;
			r.teardown = () => {
				o(), c?.();
			};
		};
	}), e;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var di = !1;
function fi(e) {
	var t = di;
	try {
		return di = !1, [e(), di];
	} finally {
		di = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
function pi(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, c = !0, l = void 0, d = () => o && i ? (l ??= /* @__PURE__ */ ut(r), W(l)) : (c && (c = !1, s = o ? sr(r) : r), s);
	let f;
	if (a) {
		var p = ce in e || ue in e;
		f = u(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = fi(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Pe(t), f(m)));
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
	var v = !1, y = (n & 1 ? ut : P)(() => (v = !1, g()));
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
var mi = "Messages for you:";
function hi(e) {
	switch (e) {
		case "pending": return "input";
		case "in_progress": return "running";
		case "completed": return "done";
		case "failed": return "failed";
		default: return e || "input";
	}
}
function gi(e) {
	if (!e.startsWith(mi)) return {
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
var _i = class {
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
		for (let e of this.blocks) e.kind === "assistant" && e.live && (e.live = !1);
	}
	apply(e) {
		let t = e.seq || 0;
		switch (e.type) {
			case "turn_start": {
				this.settle(), this.busy = !0;
				let n = gi(e.text || "");
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
			case "reasoning":
				this.busy = !0;
				break;
			case "tool": {
				let n = e.tool;
				if (!n) break;
				this.busy = !0;
				let r = n.id || n.title, i = this.tool(r);
				i ? (n.title && (i.name = n.title), n.target && (i.target = n.target), n.intent && (i.intent = n.intent), i.state = hi(n.status), n.input && (i.input = n.input), n.output && (i.output = n.output)) : (this.settle(), this.blocks.push({
					kind: "tool",
					id: r,
					name: n.title || "",
					target: n.target || "",
					intent: n.intent || "",
					input: n.input || "",
					output: n.output || "",
					state: hi(n.status),
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
}, vi = class {
	#e = /* @__PURE__ */ I("connecting");
	get status() {
		return W(this.#e);
	}
	set status(e) {
		L(this.#e, e, !0);
	}
	#t = /* @__PURE__ */ I("");
	get ended() {
		return W(this.#t);
	}
	set ended(e) {
		L(this.#t, e, !0);
	}
	#n = /* @__PURE__ */ I(0);
	get first() {
		return W(this.#n);
	}
	set first(e) {
		L(this.#n, e, !0);
	}
	#r = /* @__PURE__ */ I(!1);
	get loadingOlder() {
		return W(this.#r);
	}
	set loadingOlder(e) {
		L(this.#r, e, !0);
	}
	#i = /* @__PURE__ */ I("");
	get error() {
		return W(this.#i);
	}
	set error(e) {
		L(this.#i, e, !0);
	}
	#a = /* @__PURE__ */ I(0);
	get version() {
		return W(this.#a);
	}
	set version(e) {
		L(this.#a, e, !0);
	}
	#o = new _i();
	#s = null;
	get blocks() {
		return this.#o.blocks;
	}
	get busy() {
		return this.#o.busy;
	}
	set busy(e) {
		this.#o.busy = e;
	}
	get usage() {
		return this.#o.usage;
	}
	constructor(e, t) {
		this.api = e, this.id = t;
	}
	open(e = 200) {
		this.close(), this.#s = this.api.session.stream(this.id, { tail: e }, (e) => {
			e.type === "open" ? this.status = "live" : e.type === "end" ? (this.status = "closed", this.busy = !1) : e.type === "error" ? this.status !== "closed" && (this.status = "reconnecting") : e.type === "event" && (this.#c(e.event), e.event?.type === "state" && e.event.status === "closed" && (this.ended = e.event.resolution || "closed"));
		});
	}
	close() {
		this.#s?.(), this.#s = null;
	}
	#c(e) {
		e.seq && (!this.first || e.seq < this.first) && (this.first = e.seq), this.#o.apply(e) && this.version++;
	}
	get earlier() {
		return this.first > 1 ? this.first - 1 : 0;
	}
	async loadOlder() {
		if (!this.loadingOlder && this.earlier) {
			this.loadingOlder = !0, this.error = "";
			try {
				let e = await this.api.session.older(this.id, this.first, 400), t = new _i();
				for (let n of e.events || []) t.apply(n);
				t.settle(), this.#o.blocks = [...Be(t.blocks), ...this.#o.blocks], this.first = Math.max(1, e.first || 1), this.version++;
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
function yi() {
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
var bi = yi();
function xi(e) {
	bi = e;
}
var Si = { exec: () => null };
function Ci(e) {
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
			return i = i.replace(Ti.caret, "$1"), n = n.replace(e, i), r;
		},
		getRegex: () => new RegExp(n, t)
	};
	return r;
}
var wi = ((e = "") => {
	try {
		return !!RegExp("(?<=1)(?<!1)" + e);
	} catch {
		return !1;
	}
})(), Ti = {
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
	nextBulletRegex: Ci((e) => RegExp(`^ {0,${e}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`)),
	hrRegex: Ci((e) => RegExp(`^ {0,${e}}((?:-[ 	]*){3,}|(?:_[ 	]*){3,}|(?:\\*[ 	]*){3,})(?:\\n+|$)`)),
	fencesBeginRegex: Ci((e) => RegExp(`^ {0,${e}}(?:\`\`\`|~~~)`)),
	headingBeginRegex: Ci((e) => RegExp(`^ {0,${e}}#{1,6}(?:\\s|$)`)),
	pedanticHeadingBeginRegex: Ci((e) => RegExp(`^ {0,${e}}#`)),
	htmlBeginRegex: Ci((e) => RegExp(`^ {0,${e}}(?:</?(?:${Bi})(?: +|$|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)`, "i")),
	blockquoteBeginRegex: Ci((e) => RegExp(`^ {0,${e}}>`))
}, Ei = /^(?:[ \t]*(?:\n|$))+/, Di = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, Oi = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}(?:(?=`)\1`*|(?=~)\1~*) *(?=\n|$)|$)/, ki = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, Ai = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, ji = / {0,3}(?:[*+-]|\d{1,9}[.)])/, Mi = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |fences|blockquote|heading|hr|html|table))+?)\n {0,3}(=+|-+)[ \t]*(?:\n+|$)/, Ni = Y(Mi).replace(/bull/g, ji).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), Pi = Y(Mi).replace(/bull/g, ji).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), Fi = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table|[ \t]+\n)[^\n]+)*)/, Ii = /^[^\n]+/, Li = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\\n]|\n(?![ \t]*\n))+/, Ri = Y(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", Li).replace("title", /(?:"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\\n]|\n(?!\n))*'|\((?:\\[\s\S]|[^()\\])*\))/).getRegex(), zi = Y(/^(bull)([ \t][^\n]*?)?(?:\n|$)/).replace(/bull/g, ji).getRegex(), Bi = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", Vi = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, Hi = Y("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n*|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>[^\\n]*\\n*|$)|<![A-Z][\\s\\S]*?(?:>[^\\n]*\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>[^\\n]*\\n*|$)|</?(tag)(?:[ \\t]+|\\n|/?>|$)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!(?:script|pre|style|textarea)(?![a-z0-9-]))([a-z][a-z0-9-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</[a-z][a-z0-9-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", Vi).replace("tag", Bi).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), Ui = (e) => Y(Fi).replace("hr", ki).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", e).replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Bi).getRegex(), Wi = Ui(/ {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]/), Gi = Ui(/ {0,3}(?:[*+-]|\d{1,9}[.)])(?:[ \t]|\n|$)/), Ki = {
	blockquote: Y(/^( {0,3}> ?(?:(?![ \t]*(?:\n|$))(paragraph)|[^\n]*)(?:\n|$))+/).replace("paragraph", Gi).getRegex(),
	code: Di,
	def: Ri,
	fences: Oi,
	heading: Ai,
	hr: ki,
	html: Hi,
	lheading: Ni,
	list: zi,
	newline: Ei,
	paragraph: Wi,
	table: Si,
	text: Ii
}, qi = Y("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", ki).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Bi).getRegex(), Ji = {
	...Ki,
	lheading: Pi,
	table: qi,
	paragraph: Y(Fi).replace("hr", ki).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", qi).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]+[^ \\t\\n]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Bi).getRegex()
}, Yi = {
	...Ki,
	html: Y("^ *(?:comment(?: *\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1>(?: *\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?>(?: *\\n{2,}|\\s*$))").replace("comment", Vi).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: Si,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: Y(Fi).replace("hr", ki).replace("heading", " *#{1,6} *[^\n]").replace("lheading", Ni).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, Xi = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, Zi = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, Qi = /^( {2,}|\\)\n(?!\s*$)[ \t]*/, $i = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, ea = /[\p{P}\p{S}]/u, ta = /[\s\p{P}\p{S}]/u, na = /[^\s\p{P}\p{S}]/u, ra = Y(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, ta).getRegex(), ia = /[\p{Pi}\p{Ps}"']/u, aa = /(?!~)[\p{P}\p{S}]/u, oa = /(?!~)[\s\p{P}\p{S}]/u, sa = /(?:[^\s\p{P}\p{S}]|~)/u, ca = Y(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", wi ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), la = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, ua = Y(la, "u").replace(/punct/g, ea).getRegex(), da = Y(la, "u").replace(/punct/g, aa).getRegex(), fa = Y(/^(?:\*+(?:((?!\*)(?!openQuote)punct)|([^\s*]))?)|^_+(?:((?!_)(?!openQuote)punct)|([^\s_]))?/, "u").replace(/openQuote/g, ia).replace(/punct/g, ea).getRegex(), pa = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", ma = Y(pa, "gu").replace(/notPunctSpace/g, na).replace(/punctSpace/g, ta).replace(/punct/g, ea).getRegex(), ha = Y(pa, "gu").replace(/notPunctSpace/g, sa).replace(/punctSpace/g, oa).replace(/punct/g, aa).getRegex(), ga = Y("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)[\\s](\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|(?:(?!\\*)punct|notPunctSpace)(\\*+)(?!\\*)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, na).replace(/punctSpace/g, ta).replace(/punct/g, ea).getRegex(), _a = Y("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, na).replace(/punctSpace/g, ta).replace(/punct/g, ea).getRegex(), va = Y("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)[\\s](_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)|(?:(?!_)punct|notPunctSpace)(_+)(?!_)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, na).replace(/punctSpace/g, ta).replace(/punct/g, ea).getRegex(), ya = Y(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, ea).getRegex(), ba = Y("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, na).replace(/punctSpace/g, ta).replace(/punct/g, ea).getRegex(), xa = Y(/\\(punct)/, "gu").replace(/punct/g, ea).getRegex(), Sa = Y(/(`+)(?:[^`]|[^`][\s\S]*?[^`])\1(?!`)|`+|\\(punct)/, "gu").replace(/punct/g, ea).getRegex(), Ca = Y(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), wa = Y(Vi).replace("(?:-->|$)", "-->").getRegex(), Ta = Y("^comment|^</[a-zA-Z][a-zA-Z0-9-]*\\s*>|^<[a-zA-Z][a-zA-Z0-9-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", wa).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), Ea = /\[(?:\\[\s\S]|[^\[\]\\])*\]/, Da = Y(/(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/).replace("brackets", Ea).getRegex(), Oa = Y(/^!?\[(label)\]\([ \t\n]*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?[ \t\n]*\)/).replace("label", Da).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]+|(?=\))/).replace("title", /"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|\((?:\\[\s\S]|[^)\\])*\)/).getRegex(), ka = Y(/^!?\[(label)\]\[(ref)\]/).replace("label", Da).replace("ref", Li).getRegex(), Aa = Y(/^!?\[(ref)\](?:\[\])?/).replace("ref", Li).getRegex(), ja = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\]){1,999}/, Ma = Y(/(?:[^\[\]\\`]*(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\]))){0,999}?[^\[\]\\`]*?/).replace("brackets", Ea).getRegex(), Na = Y("reflink|nolink(?!\\()", "g").replace("reflink", Y(/^!?\[(label)\]\[(ref)\]/).replace("label", Ma).replace("ref", ja).getRegex()).replace("nolink", Y(/^!?\[(ref)\](?:\[\])?/).replace("ref", ja).getRegex()).getRegex(), Pa = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, Fa = Y(/(?:mailto:email|xmpp:email(?:\/[A-Za-z0-9@.]+)?)/).replace(/email/g, /[A-Za-z0-9._+-]+@[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(), Ia = {
	_backpedal: Si,
	anyPunctuation: xa,
	anyPunctuationOutsideCode: Sa,
	autolink: Ca,
	blockSkip: ca,
	br: Qi,
	code: Zi,
	del: Si,
	delLDelim: Si,
	delRDelim: Si,
	emStrongLDelim: ua,
	emStrongRDelimAst: ma,
	emStrongRDelimUnd: _a,
	escape: Xi,
	link: Oa,
	nolink: Aa,
	punctuation: ra,
	reflink: ka,
	reflinkSearch: Na,
	tag: Ta,
	text: $i,
	url: Si
}, La = {
	...Ia,
	emStrongLDelim: fa,
	emStrongRDelimAst: ga,
	emStrongRDelimUnd: va,
	link: Y(/^!?\[(label)\]\((.*?)\)/).replace("label", Da).getRegex(),
	reflink: Y(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", Da).getRegex()
}, Ra = {
	...Ia,
	emStrongRDelimAst: ha,
	emStrongLDelim: da,
	delLDelim: ya,
	delRDelim: ba,
	url: Y(/^emailProtocol|^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("emailProtocol", Fa).replace("protocol", Pa).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: Y(/^(?:[^a-zA-Z0-9](?=emailProtocol)|(`+|~+|[^`~])(?:(?=[`~])|(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|wwwStart|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9](?=emailProtocol)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))))/).replace("protocol", Pa).replace(/emailProtocol/g, /(?:mailto|xmpp):/).replace("wwwStart", wi ? "(?<![^\\s*_~(])www\\." : "www\\.").getRegex()
}, za = {
	...Ra,
	br: Y(Qi).replace("{2,}", "*").getRegex(),
	text: Y(Ra.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, Ba = {
	normal: Ki,
	gfm: Ji,
	pedantic: Yi
}, Va = {
	normal: Ia,
	gfm: Ra,
	breaks: za,
	pedantic: La
}, Ha = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, Ua = (e) => Ha[e];
function Wa(e, t) {
	if (t) {
		if (Ti.escapeTest.test(e)) return e.replace(Ti.escapeReplace, Ua);
	} else if (Ti.escapeTestNoEncode.test(e)) return e.replace(Ti.escapeReplaceNoEncode, Ua);
	return e;
}
function Ga(e) {
	return Ti.numericCharacterReference.test(e) ? e.replace(Ti.numericCharacterReferenceOrAmpersand, (e, t, n) => {
		if (t === void 0 && n === void 0) return "&amp;";
		let r = t === void 0 ? Number.parseInt(n, 16) : Number.parseInt(t, 10);
		return r === 0 || r > 1114111 || r >= 55296 && r <= 57343 ? "�" : r === 38 ? "&amp;" : String.fromCodePoint(r);
	}) : e;
}
function Ka(e) {
	try {
		e = encodeURI(e).replace(Ti.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function qa(e, t) {
	let n = e.replace(Ti.findPipe, (e, t, n) => {
		let r = !1, i = t;
		for (; --i >= 0 && n[i] === "\\";) r = !r;
		return r ? "|" : " |";
	}).split(Ti.splitPipe), r = 0;
	if (n[0].trim() || n.shift(), n.length > 0 && !n.at(-1)?.trim() && n.pop(), t) {
		if (n.length > t) n.splice(t);
		else for (; n.length < t;) n.push("");
	}
	for (; r < n.length; r++) n[r] = n[r].trim().replace(Ti.slashPipe, "|");
	return n;
}
function Ja(e, t, n) {
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
function Ya(e) {
	let t = e.split("\n"), n = t.length - 1;
	for (; n >= 0 && Ti.blankLine.test(t[n]);) n--;
	return t.length - n <= 2 ? e : t.slice(0, n + 1).join("\n");
}
function Xa(e) {
	return e.trim().replace(Ti.multipleSpaceGlobal, " ").toLowerCase().toUpperCase().toLowerCase();
}
function Za(e, t) {
	if (e.indexOf(t[0]) === -1 && e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
var Qa = /* @__PURE__ */ new WeakMap();
function $a(e) {
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
function eo(e, t) {
	let n = $a(t), r = Qa.get(e);
	r ? r.push(n) : Qa.set(e, [n]);
}
function to(e) {
	let t = Qa.get(e);
	t && (t.pop(), t.length === 0 && Qa.delete(e));
}
function no(e, t, n) {
	let r = Qa.get(e)?.at(-1);
	return r && r.maskedSrc === t ? r.lastPossibleEmStrongCloserIndex[n] : $a(t).lastPossibleEmStrongCloserIndex[n];
}
function ro(e, t) {
	let n = Qa.get(e)?.at(-1);
	return n && n.maskedSrc === t ? n.lastPossibleDelCloserIndex : $a(t).lastPossibleDelCloserIndex;
}
function io(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function ao(e, t, n, r, i) {
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
function oo(e, t, n) {
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
function so(e, t, n, r) {
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
var co = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || bi;
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
			let e = this.options.pedantic ? t[0] : Ya(t[0]);
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
			let e = t[0], n = oo(e, t[3] || "", this.rules);
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
				let t = Ja(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceTabChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: Ja(t[0], "\n"),
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
			raw: Ja(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = Ja(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteMarkerWhitespace, (e, t, n) => t + io(n, t.length)).replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
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
				let u = t[2].split("\n", 1)[0], d = t[1].length, f = this.options.pedantic ? io(u, d) : u.replace(this.rules.other.leadingSpaceTab, (e) => io(e, d)), p = e.split("\n", 1)[0], m = !f.trim(), h = 0;
				if (this.options.pedantic ? (h = 2, a = f.trimStart()) : m ? h = d + 1 : (h = f.search(this.rules.other.nonSpaceChar), h = h > 4 ? 1 : h, a = f.slice(h), h += d), m && this.rules.other.blankLine.test(p) && (e = e.substring(p.length + 1), i = !0), !i) {
					let t = this.rules.other.nextBulletRegex(h), n = this.rules.other.hrRegex(h), r = this.rules.other.fencesBeginRegex(h), i = this.options.pedantic ? this.rules.other.pedanticHeadingBeginRegex(h) : this.rules.other.headingBeginRegex(h), o = this.rules.other.htmlBeginRegex(h), s = this.rules.other.blockquoteBeginRegex(h);
					for (; e;) {
						let c = e.split("\n", 1)[0], l;
						if (p = c, this.options.pedantic ? (p = p.replace(this.rules.other.listReplaceNesting, "  "), l = p) : l = p.replace(this.rules.other.leadingSpaceTab, (e) => io(e)), r.test(p) || i.test(p) || o.test(p) || s.test(p) || t.test(p) || n.test(p)) break;
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
			let e = Ya(t[0]);
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
			if (!this.rules.other.startAngleBracket.test(t[2]) && Za(t[2], "()") !== -1) return;
			let e = Xa(t[1]), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: Ja(t[0], "\n"),
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = qa(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: Ja(t[0], "\n"),
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
			for (let e of i) a.rows.push(qa(e, a.header.length).map((e, t) => ({
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
				raw: Ja(t[0], "\n"),
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
			if (!this.options.pedantic && so(e, t[1], n, this.rules)) return;
			let r = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(r)) {
				if (!this.rules.other.endAngleBracket.test(r)) return;
				let e = Ja(r.slice(0, -1), "\\");
				if ((r.length - e.length) % 2 == 0) return;
			} else {
				let e = Za(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let i = Ja(t[0].slice(0, -1), "\\");
			if ((t[0].length - i.length) % 2 == 0) return;
			let a = t[2], o = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(a);
				e && (a = e[1], o = e[3]);
			} else o = t[3] ? t[3].slice(1, -1) : "";
			return a = a.trim(), this.rules.other.startAngleBracket.test(a) && (a = this.options.pedantic && !this.rules.other.endAngleBracket.test(r) ? a.slice(1) : a.slice(1, -1)), ao(t, {
				href: a && a.replace(this.rules.inline.anyPunctuation, "$1"),
				title: o && o.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let r = n[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && so(e, n[1], r, this.rules)) return;
			let i = t[Xa(n[2] || n[1])];
			if (!i) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return ao(n, i, n[0], this.lexer, this.rules);
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
			let _ = i === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd, v = no(this, t, i), y = t.length - e.length + n;
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
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim, c = ro(this, t), l = t.length - e.length + n;
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
				text: e ? t[0] : Ga(t[0]),
				escaped: e
			};
		}
	}
}, lo = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || bi, this.options.tokenizer = this.options.tokenizer || new co(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			linkEmitted: !1,
			linkParenPossible: !0,
			top: !0
		};
		let t = {
			other: Ti,
			block: Ba.normal,
			inline: Va.normal
		};
		this.options.pedantic ? (t.block = Ba.pedantic, t.inline = Va.pedantic) : this.options.gfm && (t.block = Ba.gfm, t.inline = this.options.breaks ? Va.breaks : Va.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: Ba,
			inline: Va
		};
	}
	static lex(t, n) {
		return new e(n).lex(t);
	}
	static lexInline(t, n) {
		return new e(n).inlineTokens(t);
	}
	lex(e) {
		e = e.replace(Ti.carriageReturn, "\n"), this.blockTokens(e, this.tokens);
		for (let e = 0; e < this.inlineQueue.length; e++) {
			let t = this.inlineQueue[e];
			this.state.inRawBlock = !1, this.inlineTokens(t.src, t.tokens);
		}
		return this.inlineQueue = [], this.tokens;
	}
	blockTokens(e, t = [], n = !1) {
		this.tokenizer.lexer = this, this.options.pedantic && (e = e.replace(Ti.tabCharGlobal, "    ").replace(Ti.spaceLine, ""));
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
			if (e.charAt(0) !== "!" && Object.hasOwn(this.tokens.links, Xa(e.slice(n + 1, -1))) && !(n > 1 && this.linkInText(e.slice(1, n - 1)))) return !0;
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
		let n = e.includes("*") || e.includes("_") || e.includes("~") || this.options.hooks?.emStrongMask || this.tokenizer.emStrong !== co.prototype.emStrong || this.tokenizer.del !== co.prototype.del ? this.#n(e) : e;
		eo(this.tokenizer, n);
		try {
			return this.#t(e, n, t);
		} finally {
			to(this.tokenizer);
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
				if (!Object.hasOwn(this.tokens.links, Xa(t.slice(r + 1, -1)))) return t;
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
}, uo = class {
	options;
	parser;
	constructor(e) {
		this.options = e || bi;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(Ti.notSpaceStart)?.[0], i = e ? e.replace(Ti.endingNewline, "") + "\n" : "";
		return r ? "<pre><code class=\"language-" + Wa(r) + "\">" + (n ? i : Wa(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : Wa(i, !0)) + "</code></pre>\n";
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
		return `<code>${Wa(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, text: n, tokens: r, autolink: i }) {
		let a = i ? Wa(n, !0) : this.parser.parseInline(r), o = Ka(e);
		if (o === null) return a;
		e = Wa(o, i);
		let s = "<a href=\"" + e + "\"";
		return t && (s += " title=\"" + Wa(t) + "\""), s += ">" + a + "</a>", s;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = Ka(e);
		if (i === null) return Wa(n);
		e = i;
		let a = `<img src="${Wa(e)}" alt="${Wa(n)}"`;
		return t && (a += ` title="${Wa(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : Wa(e.text);
	}
}, fo = class {
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
}, po = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || bi, this.options.renderer = this.options.renderer || new uo(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new fo(), this.textRenderer.parser = this;
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
}, mo = class {
	options;
	block;
	constructor(e) {
		this.options = e || bi;
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
		return e ? lo.lex : lo.lexInline;
	}
	provideParser(e = this.block) {
		return e ? po.parse : po.parseInline;
	}
};
function ho(e, t) {
	if (Array.isArray(t)) for (let n = 0; n < t.length; n++) e.push(t[n]);
	else e.push(t);
}
var go = new class {
	defaults = yi();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = po;
	Renderer = uo;
	TextRenderer = fo;
	Lexer = lo;
	Tokenizer = co;
	Hooks = mo;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (ho(n, t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) ho(n, this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) ho(n, this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				ho(n, this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					ho(n, this.walkTokens(i, t));
				}) : e.tokens && ho(n, this.walkTokens(e.tokens, t));
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
				let t = this.defaults.renderer || new uo(this.defaults);
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
				let t = this.defaults.tokenizer || new co(this.defaults);
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
				let t = this.defaults.hooks || new mo();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = mo.passThroughHooks.has(n) ? (e) => {
						if (t.options.async && mo.passThroughHooksRespectAsync.has(n)) return (async () => {
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
					return ho(n, r.call(this, e)), t && ho(n, t.call(this, e)), n;
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
		return lo.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return po.parse(e, t ?? this.defaults);
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
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? lo.lex : lo.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? po.parse : po.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? lo.lex : lo.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? po.parse : po.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + Wa(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function X(e, t) {
	return go.parse(e, t);
}
X.options = X.setOptions = function(e) {
	return go.setOptions(e), X.defaults = go.defaults, xi(X.defaults), X;
}, X.getDefaults = yi, X.defaults = bi;
function _o(...e) {
	return go.use(...e), X.defaults = go.defaults, xi(X.defaults), X;
}
X.use = _o, X.walkTokens = function(e, t) {
	return go.walkTokens(e, t);
}, X.parseInline = go.parseInline, X.Parser = po, X.parser = po.parse, X.Renderer = uo, X.TextRenderer = fo, X.Lexer = lo, X.lexer = lo.lex, X.Tokenizer = co, X.Hooks = mo, X.parse = X, X.options, X.setOptions, X.walkTokens, X.parseInline, po.parse, lo.lex;
// node_modules/dompurify/dist/purify.es.mjs
function vo(e, t) {
	this.v = e, this.k = t;
}
function yo(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function bo(e) {
	if (Array.isArray(e)) return e;
}
function xo(e, t) {
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
function So() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function Co(e, t) {
	return bo(e) || xo(e, t) || wo(e, t) || So();
}
function wo(e, t) {
	if (e) {
		if (typeof e == "string") return yo(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? yo(e, t) : void 0;
	}
}
function To(e) {
	var t, n;
	function r(t, n) {
		try {
			var a = e[t](n), o = a.value, s = o instanceof vo;
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
To.prototype[typeof Symbol == "function" && Symbol.asyncIterator || "@@asyncIterator"] = function() {
	return this;
}, To.prototype.next = function(e) {
	return this._invoke("next", e);
}, To.prototype.throw = function(e) {
	return this._invoke("throw", e);
}, To.prototype.return = function(e) {
	return this._invoke("return", e);
};
var Eo = Object.entries, Do = Object.setPrototypeOf, Oo = Object.isFrozen, ko = Object.getPrototypeOf, Ao = Object.getOwnPropertyDescriptor, jo = Object.freeze, Mo = Object.seal, No = Object.create, Po = typeof Reflect < "u" && Reflect, Fo = Po.apply, Io = Po.construct;
jo ||= function(e) {
	return e;
}, Mo ||= function(e) {
	return e;
}, Fo ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, Io ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var Lo = Z(Array.prototype.forEach);
Array.prototype.indexOf;
var Ro = Z(Array.prototype.lastIndexOf), zo = Z(Array.prototype.pop), Bo = Z(Array.prototype.push), Vo = Z(Array.prototype.slice), Ho = Z(Array.prototype.splice), Uo = Array.isArray, Wo = ss(Set), Go = Z(Set.prototype.add), Ko = Z(Set.prototype.has), qo = Z(String.prototype.toLowerCase), Jo = Z(String.prototype.toString), Yo = Z(String.prototype.match), Xo = Z(String.prototype.replace), Zo = Z(String.prototype.indexOf), Qo = Z(String.prototype.trim), $o = Z(Number.prototype.toString), es = Z(Boolean.prototype.toString), ts = typeof BigInt > "u" ? null : Z(BigInt.prototype.toString), ns = typeof Symbol > "u" ? null : Z(Symbol.prototype.toString), rs = Z(Object.prototype.hasOwnProperty), is = Z(Object.prototype.toString), as = Z(RegExp.prototype.test), os = ss(TypeError);
function Z(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return Fo(e, t, n);
	};
}
function ss(e) {
	return function() {
		return Io(e, [...arguments]);
	};
}
function Q(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : qo;
	if (Do && Do(e, null), !Uo(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (Oo(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function cs(e) {
	for (let t = 0; t < e.length; t++) rs(e, t) || (e[t] = null);
	return e;
}
function ls(e) {
	let t = No(null);
	for (let r of Eo(e)) {
		var n = Co(r, 2);
		let i = n[0], a = n[1];
		rs(e, i) && (t[i] = Uo(a) ? cs(a) : a && typeof a == "object" && a.constructor === Object ? ls(a) : a);
	}
	return t;
}
function us(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return $o(e);
		case "boolean": return es(e);
		case "bigint": return ts ? ts(e) : "0";
		case "symbol": return ns ? ns(e) : "Symbol()";
		case "undefined": return is(e);
		case "function":
		case "object": {
			if (e === null) return is(e);
			let t = e, n = ds(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : is(e);
			}
			return is(e);
		}
		default: return is(e);
	}
}
function ds(e, t) {
	for (; e !== null;) {
		let n = Ao(e, t);
		if (n) {
			if (n.get) return Z(n.get);
			if (typeof n.value == "function") return Z(n.value);
		}
		e = ko(e);
	}
	function n() {
		return null;
	}
	return n;
}
function fs(e, t) {
	for (; e !== null;) {
		let n = Ao(e, t);
		if (n) return n.get ? Z(n.get) : null;
		e = ko(e);
	}
	return null;
}
function ps(e) {
	try {
		return as(e, ""), !0;
	} catch {
		return !1;
	}
}
var ms = jo(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), hs = jo(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), gs = jo([
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
]), _s = jo([
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
]), vs = jo(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), ys = jo([
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
]), bs = jo(["#text"]), xs = jo(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), Ss = jo(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dominant-baseline.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.pointer-events.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-orientation.text-rendering.textlength.type.u1.u2.unicode.values.vector-effect.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), Cs = jo(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), ws = jo([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), Ts = Mo(/{{[\w\W]*|^[\w\W]*}}/g), Es = Mo(/<%[\w\W]*|^[\w\W]*%>/g), Ds = Mo(/\${[\w\W]*/g), Os = Mo(/^data-[\-\w.\u00B7-\uFFFF]+$/), ks = Mo(/^aria-[\-\w]+$/), As = Mo(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), js = Mo(/^(?:\w+script|data):/i), Ms = Mo(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), Ns = Mo(/^html$/i), Ps = Mo(/^[a-z][.\w]*(-[.\w]+)+$/i), Fs = Mo(/<[/\w!]/g), Is = Mo(/<[/\w]/g), Ls = Mo(/<\/no(script|embed|frames)/i), Rs = Mo(/\/>/i), $ = {
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
}, zs = jo(Q({}, [
	"style",
	"script",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"plaintext",
	"noscript"
])), Bs = function() {
	return typeof window > "u" ? null : window;
}, Vs = function(e, t) {
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
}, Hs = function() {
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
}, Us = function(e, t, n, r) {
	return rs(e, t) && Uo(e[t]) ? Q(r.base ? ls(r.base) : {}, e[t], r.transform) : n;
}, Ws = function(e, t, n) {
	let r = rs(e, t) ? e[t] : void 0;
	return r && typeof r == "object" ? ls(r) : n();
};
function Gs() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : Bs(), t = (e) => Gs(e);
	if (t.version = "3.4.17", t.removed = [], !e || !e.document || e.document.nodeType !== $.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = ds(d, "cloneNode"), p = ds(d, "remove"), m = ds(d, "removeAttributeNode"), h = ds(d, "nextSibling"), g = ds(d, "childNodes"), _ = ds(d, "parentNode"), v = ds(d, "shadowRoot"), y = ds(d, "attributes"), b = o && o.prototype ? ds(o.prototype, "nodeType") : null, x = o && o.prototype ? ds(o.prototype, "nodeName") : null, S = o && o.prototype ? ds(o.prototype, "ownerDocument") : null, C = function(e) {
		return b ? b(e) : e.nodeType;
	}, w = function(e) {
		return x ? x(e) : e.nodeName;
	}, ee = fs(d, "namespaceURI"), te = function(e) {
		return ee ? ee(e) : e.namespaceURI;
	};
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let T, ne = "", re, ie = !1, ae = 0, oe = function() {
		if (ae > 0) throw os("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
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
		return ie ||= (re = Vs(u, i), !0), re;
	}, ue = n, de = ue.implementation, fe = ue.createNodeIterator, pe = ue.createDocumentFragment, me = ue.getElementsByTagName, he = r.importNode, E = Hs();
	t.isSupported = typeof Eo == "function" && typeof _ == "function" && de && de.createHTMLDocument !== void 0;
	let ge = Ts, _e = Es, ve = Ds, ye = Os, be = ks, D = js, xe = Ms, O = Ps, Se = As, k = null, A = Q({}, [
		...ms,
		...hs,
		...gs,
		...vs,
		...bs
	]), j = null, Ce = Q({}, [
		...xs,
		...Ss,
		...Cs,
		...ws
	]), we = Object.seal(No(null, {
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
	})), Te = null, Ee = null, De = Object.seal(No(null, {
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
	})), Oe = !0, ke = !0, Ae = !1, je = !0, Me = !1, Ne = !0, Pe = !1, Fe = !1, Ie = null, Le = null, Re = !1, ze = !1, Be = !1, Ve = !1, He = !0, Ue = !1, We = "user-content-", Ge = !0, Ke = !1, qe = null, Je = {}, Ye = null, Xe = Q({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Ze = null, Qe = Q({}, [
		"audio",
		"video",
		"img",
		"source",
		"image",
		"track"
	]), M = null, $e = Q({}, [
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
	], Jo), st = jo([
		"mi",
		"mo",
		"mn",
		"ms",
		"mtext"
	]), ct = Q({}, st), lt = jo(["annotation-xml"]), ut = Q({}, lt), dt = Q({}, [
		"title",
		"style",
		"font",
		"a",
		"script"
	]), ft = null, N = ["application/xhtml+xml", "text/html"], P = null, pt = null, mt = n.createElement("form"), ht = function(e) {
		return e instanceof RegExp || e instanceof Function;
	}, gt = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		if (pt && pt === e) return;
		e && typeof e == "object" || (e = {}), e = ls(e), ft = N.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, P = ft === "application/xhtml+xml" ? Jo : qo, k = Us(e, "ALLOWED_TAGS", A, { transform: P }), j = Us(e, "ALLOWED_ATTR", Ce, { transform: P }), at = Us(e, "ALLOWED_NAMESPACES", ot, { transform: Jo }), M = Us(e, "ADD_URI_SAFE_ATTR", $e, {
			transform: P,
			base: $e
		}), Ze = Us(e, "ADD_DATA_URI_TAGS", Qe, {
			transform: P,
			base: Qe
		}), Ye = Us(e, "FORBID_CONTENTS", Xe, { transform: P }), Te = Us(e, "FORBID_TAGS", ls({}), { transform: P }), Ee = Us(e, "FORBID_ATTR", ls({}), { transform: P }), Je = rs(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? ls(e.USE_PROFILES) : e.USE_PROFILES : !1, Oe = e.ALLOW_ARIA_ATTR !== !1, ke = e.ALLOW_DATA_ATTR !== !1, Ae = e.ALLOW_UNKNOWN_PROTOCOLS || !1, je = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, Me = e.SAFE_FOR_TEMPLATES || !1, Ne = e.SAFE_FOR_XML !== !1, Pe = e.WHOLE_DOCUMENT || !1, ze = e.RETURN_DOM || !1, Be = e.RETURN_DOM_FRAGMENT || !1, Ve = e.RETURN_TRUSTED_TYPE || !1, Re = e.FORCE_BODY || !1, He = e.SANITIZE_DOM !== !1, Ue = e.SANITIZE_NAMED_PROPS || !1, Ge = e.KEEP_CONTENT !== !1, Ke = e.IN_PLACE || !1, Se = ps(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : As, rt = typeof e.NAMESPACE == "string" ? e.NAMESPACE : nt, ct = Ws(e, "MATHML_TEXT_INTEGRATION_POINTS", () => Q({}, st)), ut = Ws(e, "HTML_INTEGRATION_POINTS", () => Q({}, lt));
		let t = Ws(e, "CUSTOM_ELEMENT_HANDLING", () => No(null));
		if (we = No(null), rs(t, "tagNameCheck") && ht(t.tagNameCheck) && (we.tagNameCheck = t.tagNameCheck), rs(t, "attributeNameCheck") && ht(t.attributeNameCheck) && (we.attributeNameCheck = t.attributeNameCheck), rs(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (we.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), Mo(we), Me && (ke = !1), Be && (ze = !0), Je && (k = Q({}, bs), j = No(null), Je.html === !0 && (Q(k, ms), Q(j, xs)), Je.svg === !0 && (Q(k, hs), Q(j, Ss), Q(j, ws)), Je.svgFilters === !0 && (Q(k, gs), Q(j, Ss), Q(j, ws)), Je.mathMl === !0 && (Q(k, vs), Q(j, Cs), Q(j, ws))), De.tagCheck = null, De.attributeCheck = null, rs(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? De.tagCheck = e.ADD_TAGS : Uo(e.ADD_TAGS) && (k === A && (k = ls(k)), Q(k, e.ADD_TAGS, P))), rs(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? De.attributeCheck = e.ADD_ATTR : Uo(e.ADD_ATTR) && (j === Ce && (j = ls(j)), Q(j, e.ADD_ATTR, P))), rs(e, "ADD_FORBID_CONTENTS") && Uo(e.ADD_FORBID_CONTENTS) && (Ye === Xe && (Ye = ls(Ye)), Q(Ye, e.ADD_FORBID_CONTENTS, P)), Ge && (k["#text"] = !0), Pe && Q(k, [
			"html",
			"head",
			"body"
		]), k.table && (Q(k, ["tbody"]), delete Te.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw os("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw os("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
			let t = T;
			T = e.TRUSTED_TYPES_POLICY;
			try {
				ne = se("");
			} catch (e) {
				throw T = t, e;
			}
		} else e.TRUSTED_TYPES_POLICY === null ? (T = void 0, ne = "") : (T === void 0 && (T = le()), T && typeof ne == "string" && (ne = se("")));
		jo && jo(e), pt = e;
	}, _t = Q({}, [
		...hs,
		...gs,
		..._s
	]), vt = Q({}, [...vs, ...ys]), F = function(e, t, n) {
		return t.namespaceURI === nt ? e === "svg" : t.namespaceURI === et ? e === "svg" && (n === "annotation-xml" || ct[n]) : !!_t[e];
	}, yt = function(e, t, n) {
		return t.namespaceURI === nt ? e === "math" : t.namespaceURI === tt ? e === "math" && ut[n] : !!vt[e];
	}, bt = function(e, t, n) {
		return t.namespaceURI === tt && !ut[n] || t.namespaceURI === et && !ct[n] ? !1 : !vt[e] && (dt[e] || !_t[e]);
	}, xt = function(e) {
		let t = te(e), n = qo(w(e)), r = _(e), i, a;
		return r && C(r) === $.element ? (i = { namespaceURI: te(r) }, a = qo(w(r))) : (i = { namespaceURI: rt }, a = "template"), at[t] ? t === tt ? F(n, i, a) : t === et ? yt(n, i, a) : t === nt ? bt(n, i, a) : !!(ft === "application/xhtml+xml" && at[t]) : !1;
	}, St = function(e) {
		Bo(t.removed, { element: e });
		try {
			_(e).removeChild(e);
		} catch {
			if (p(e), !_(e)) throw os("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
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
		t && Lo(Vo(t), (e) => {
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
		Bo(t.removed, {
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
			typeof i != "string" || j[P(i)] || Ct(e, r, i);
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
		let n = Wo();
		Lo(kt(e), (e) => {
			C(e) === $.element && Go(n, e);
		}), Lo(t, (e) => {
			C(e) === $.element && !Ko(n, e) && Et(e);
		});
	}, jt = function(e, n) {
		let r = !1;
		return Lo(t.removed, (t) => {
			t.element && (r ||= t.element === e, Ot(t.element));
		}), At(e, n), r;
	}, Mt = function(e, t) {
		let n = qe, r = null, i = () => (r || (r = Wo(), Lo(t, (e) => {
			Go(r, e);
		})), r);
		for (let t = 0;; t++) {
			let r = kt(e), a = [e], o = !1;
			if (Lo(r, (e) => {
				C(e) === $.documentFragment ? Bo(a, e) : !o && !Ko(n, e) && (o = Ko(i(), e));
			}), !o) return;
			if (t >= 8) throw os("the tree kept changing during in-place sanitization; refusing to return it");
			Lo(a, (t, n) => {
				Nt(t, n > 0, e, i());
			});
		}
	}, Nt = function(e, t, n, r) {
		let i = Rt(e), a;
		for (; a = i.nextNode();) Ko(r, a) && !Ko(qe, a) && (t && Ut(E.uponSanitizeShadowNode, a, null), Yt(a, n), R(a, n), Vt(a.content) && nn(a.content));
	}, Pt = function(e) {
		qe && Go(qe, e);
	}, Ft = function(e, t) {
		return Ne ? e === "patchsrc" || e === "for" && t !== "label" && t !== "output" : !1;
	}, It = function(e) {
		Ne && Dt(e, (e) => {
			let t = C(e);
			if (t === $.processingInstruction || t === $.comment && as(Is, e.data)) {
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
		if (Re) e = "<remove></remove>" + e;
		else {
			let t = Yo(e, /^[\r\n\t ]+/);
			r = t && t[0];
		}
		ft === "application/xhtml+xml" && rt === nt && (e = "<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>" + e + "</body></html>");
		let i = T ? se(e) : e;
		if (rt === nt) try {
			t = new l().parseFromString(i, ft);
		} catch {}
		if (!t || !t.documentElement) {
			t = de.createDocument(rt, "template", null);
			try {
				t.documentElement.innerHTML = it ? ne : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), rt === nt ? me.call(t, Pe ? "html" : "body")[0] : Pe ? t.documentElement : a;
	}, Rt = function(e) {
		let t = S ? S(e) : e.ownerDocument;
		return fe.call(t || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, zt = function(e) {
		return e = Xo(e, ge, " "), e = Xo(e, _e, " "), e = Xo(e, ve, " "), e;
	}, I = function(e) {
		Bt(e);
		let t = e.querySelectorAll?.call(e, "template");
		t && Lo(t, (e) => {
			Vt(e.content) && I(e.content);
		});
	}, Bt = function(e) {
		e.normalize();
		let t = S ? S(e) : e.ownerDocument, n = fe.call(t || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), r = n.nextNode();
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
		e.length !== 0 && Lo(e, (e) => {
			e.call(t, n, r, pt);
		});
	}
	let Wt = function(e, t) {
		return !!(Ne && e.hasChildNodes() && !Ht(e.firstElementChild) && as(Fs, e.textContent) && as(Fs, e.innerHTML) || Ne && zs[t] && C(e) === $.element && te(e) === nt && (Ht(e.firstElementChild) || typeof e.textContent == "string" && as(Fs, e.textContent)) || e.nodeType === $.processingInstruction || Ne && e.nodeType === $.comment && as(Is, e.data));
	}, Gt = function(e, t) {
		return e instanceof RegExp ? as(e, t) : e instanceof Function && !!e(t, ...[...arguments].slice(2));
	}, Kt = function(e, t, n) {
		if (!Te[t] && Qt(t) && Gt(we.tagNameCheck, t)) return !1;
		if (Ge && !Ye[t] && !Ye[qo(t)]) {
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
		return e.length === 0 ? t : t === n || t === r ? ls(t) : t;
	}, Jt = function(e, t) {
		return e === t || _(e) !== null ? !1 : (Ke && Ot(e), !0);
	}, Yt = function(e, n) {
		if (Pt(e), Ut(E.beforeSanitizeElements, e, null), Jt(e, n)) return !0;
		if (L(e)) return St(e), !0;
		let r = P(w(e));
		if (k = qt(E.uponSanitizeElement, k, A, Ie), Ut(E.uponSanitizeElement, e, {
			tagName: r,
			allowedTags: k
		}), Jt(e, n)) return !0;
		if (Wt(e, r)) return St(e), !0;
		if (Te[r] || !(De.tagCheck instanceof Function && De.tagCheck(r)) && !k[r]) {
			let t = Kt(e, r, n);
			return t === !1 && (Ut(E.afterSanitizeElements, e, null), Jt(e, n)) ? !0 : t;
		}
		if (C(e) === $.element && !xt(e) || (r === "noscript" || r === "noembed" || r === "noframes") && as(Ls, e.innerHTML)) return St(e), !0;
		if (Me && e.nodeType === $.text) {
			let n = zt(e.textContent);
			e.textContent !== n && (Bo(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return Ut(E.afterSanitizeElements, e, null), Jt(e, n);
	}, Xt = function(e, t, r) {
		if (Ee[t] || Ft(t, e) || He && (t === "id" || t === "name") && (r in n || r in mt)) return !1;
		let i = j[t] || De.attributeCheck instanceof Function && De.attributeCheck(t, e);
		return ke && as(ye, t) || Oe && as(be, t) ? !0 : i ? M[t] || as(Se, Xo(r, xe, "")) || (t === "src" || t === "xlink:href" || t === "href") && e !== "script" && Zo(r, "data:") === 0 && Ze[e] || Ae && !as(D, Xo(r, xe, "")) ? !0 : !r : Qt(e) && Gt(we.tagNameCheck, e) && Gt(we.attributeNameCheck, t, e) || t === "is" && we.allowCustomizedBuiltInElements && Gt(we.tagNameCheck, r);
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
		return !Zt[qo(e)] && as(O, e);
	}, $t = function(e, t, n, r) {
		if (T && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return se(r);
			case "TrustedScriptURL": return ce(r);
		}
		return r;
	}, en = function(e, t, n, r) {
		try {
			return n ? e.setAttributeNS(n, t, r) : e.setAttribute(t, r), !L(e) || (St(e), !1);
		} catch {
			return Tt(t, e), !1;
		}
	}, tn = function(e, t) {
		return Ne && as(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, t) ? !0 : qo(e) === "attributename" && Yo(t, "href") !== null;
	}, R = function(e, n) {
		if (Ut(E.beforeSanitizeAttributes, e, null), Jt(e, n)) return;
		let r = e.attributes;
		if (!r || L(e)) return;
		j = qt(E.uponSanitizeAttribute, j, Ce, Le);
		let i = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: j,
			forceKeepAttr: void 0
		}, a = r.length, o = P(e.nodeName);
		for (; a--;) {
			let n = r[a], s = n.name, c = n.namespaceURI, l = n.value, u = P(s), d = l, f = s === "value" ? d : Qo(d), p = !1;
			i.attrName = u, i.attrValue = f, i.keepAttr = !0, i.forceKeepAttr = void 0, Ut(E.uponSanitizeAttribute, e, i), f = i.attrValue, Ue && (u === "id" || u === "name") && Zo(f, We) !== 0 && (Tt(s, e, n), f = We + f, p = !0), tn(s, f) ? Tt(s, e, n) : i.forceKeepAttr || (i.keepAttr ? !je && as(Rs, f) ? Tt(s, e, n) : (Me && (f = zt(f)), Xt(o, u, f) ? (f = $t(o, u, c, f), f !== d && en(e, s, c, f) && p && zo(t.removed)) : Tt(s, e, n)) : Tt(s, e, n));
		}
		Ut(E.afterSanitizeAttributes, e, null), Jt(e, n);
	}, nn = function(e) {
		let t = null, n = Rt(e);
		for (Ut(E.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (Ut(E.uponSanitizeShadowNode, t, null), Yt(t, e), R(t, e), Vt(t.content) && nn(t.content), C(t) === $.element) {
			let e = v(t);
			Vt(e) && (z(e), nn(e));
		}
		Me && Bt(e), Ut(E.afterSanitizeShadowDOM, e, null);
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
		if (it = !e, it && (e = "<!-->"), typeof e != "string" && !Ht(e) && (e = us(e), typeof e != "string")) throw os("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		Fe ? (k = Ie, j = Le) : gt(n), (E.uponSanitizeElement.length > 0 || E.uponSanitizeAttribute.length > 0) && (k = ls(k)), E.uponSanitizeAttribute.length > 0 && (j = ls(j)), t.removed = [];
		let c = Ke && typeof e != "string" && Ht(e), l = [], u = qe;
		if (c) {
			It(e);
			let t = w(e);
			if (typeof t == "string") {
				let n = P(t);
				if (!k[n] || Te[n]) throw wt(e), os("root node is forbidden and cannot be sanitized in-place");
			}
			if (L(e)) throw wt(e), os("root node is clobbered and cannot be sanitized in-place");
			l = kt(e), qe = Wo();
		} else if (Ht(e)) i = Lt("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === $.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), z(i);
		else {
			if (!ze && !Me && !Pe && e.indexOf("<") === -1) return T && Ve ? se(e) : e;
			if (i = Lt(e), !i) return ze ? null : Ve ? ne : "";
		}
		i && Re && St(i.firstChild);
		let d = c ? e : i;
		try {
			c && z(e);
			let t = Rt(d);
			for (; o = t.nextNode();) Yt(o, d), R(o, d), Vt(o.content) && nn(o.content);
			c && Mt(e, l);
		} catch (t) {
			throw c && (wt(e), jt(e, l)), t;
		} finally {
			qe = u;
		}
		if (c) {
			if (jt(e, l)) throw os("a node selected for removal could not be safely returned; refusing to sanitize in place");
			return Me && I(e), e;
		}
		if (ze) {
			if (Me && I(i), Be) for (s = pe.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (j.shadowroot || j.shadowrootmode) && (s = he.call(r, s, !0)), s;
		}
		let f = Pe ? i.outerHTML : i.innerHTML;
		return Pe && k["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && as(Ns, i.ownerDocument.doctype.name) && (f = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + f), Me && (f = zt(f)), T && Ve ? se(f) : f;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		gt(e), Fe = !0, Ie = k, Le = j;
	}, t.clearConfig = function() {
		pt = null, Fe = !1, Ie = null, Le = null, T = re, ne = "";
	}, t.isValidAttribute = function(e, t, n) {
		pt || gt({});
		let r = P(e), i = P(t);
		return typeof n == "string" && (tn(t, n) || !je && as(Rs, n)) ? !1 : Xt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && rs(E, e) && Bo(E[e], t);
	}, t.removeHook = function(e, t) {
		if (rs(E, e)) {
			if (t !== void 0) {
				let n = Ro(E[e], t);
				return n === -1 ? void 0 : Ho(E[e], n, 1)[0];
			}
			return zo(E[e]);
		}
	}, t.removeHooks = function(e) {
		rs(E, e) && (E[e] = []);
	}, t.removeAllHooks = function() {
		E = Hs();
	}, t;
}
var Ks = Gs();
// plugins/chat/format.js
function qs(e) {
	let t = X.parse(e || "", {
		async: !1,
		gfm: !0,
		breaks: !1
	}), n = Ks.sanitize(t, {
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
function Js(e) {
	return e >= 1e6 ? (e / 1e6).toFixed(1) + "M" : e >= 1e3 ? Math.round(e / 1e3) + "k" : String(e);
}
var Ys = (e) => "desk-chat-prompts." + e;
function Xs(e, t) {
	let n = [];
	try {
		let t = JSON.parse(localStorage.getItem(Ys(e)) ?? "[]");
		Array.isArray(t) && (n = t.filter((e) => typeof e == "string"));
	} catch {}
	return [...new Set([...t, ...n.filter((e) => !t.includes(e))].filter(Boolean).reverse())].slice(0, 100);
}
function Zs(e, t) {
	try {
		let n = Xs(e, []).filter((e) => e !== t);
		localStorage.setItem(Ys(e), JSON.stringify([...n.reverse(), t].slice(-100)));
	} catch {}
}
// kit/format.js
function Qs(e, t) {
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
// plugins/chat/Block.svelte
var $s = /* @__PURE__ */ G("<details class=\"prompt svelte-5n3hzh\"><summary class=\"svelte-5n3hzh\"> </summary><pre class=\"svelte-5n3hzh\"> </pre></details>"), ec = /* @__PURE__ */ G("<div class=\"md svelte-5n3hzh\"></div>"), tc = /* @__PURE__ */ G("<div class=\"user svelte-5n3hzh\" data-testid=\"block-user\"><span class=\"tag svelte-5n3hzh\"> </span> <!></div>"), nc = /* @__PURE__ */ G("<div data-testid=\"block-assistant\"></div>"), rc = /* @__PURE__ */ G("<pre class=\"io svelte-5n3hzh\"> </pre>"), ic = /* @__PURE__ */ G("<pre class=\"io out svelte-5n3hzh\"> </pre>"), ac = /* @__PURE__ */ G("<!> <!>", 1), oc = /* @__PURE__ */ G("<div class=\"tool svelte-5n3hzh\" data-testid=\"block-tool\"><button class=\"tool-head svelte-5n3hzh\"><span class=\"tname svelte-5n3hzh\" data-testid=\"tool-label\"> </span> <span class=\"tintent svelte-5n3hzh\" data-testid=\"tool-detail\"> </span> <span class=\"tstate svelte-5n3hzh\" data-testid=\"tool-state\"> </span></button> <!></div>"), sc = /* @__PURE__ */ G("<p class=\"error svelte-5n3hzh\" data-testid=\"block-error\"> </p>"), cc = /* @__PURE__ */ G("<pre class=\"svelte-5n3hzh\"> </pre>"), lc = /* @__PURE__ */ G("<details class=\"system svelte-5n3hzh\" data-testid=\"block-system\"><summary class=\"svelte-5n3hzh\"> </summary> <!></details>");
function uc(e, t) {
	We(t, !0);
	let n = pi(t, "agent", 3, null), r = /* @__PURE__ */ I(!1);
	function i(e, t = 140) {
		let n = (e || "").replace(/\s+/g, " ").trim();
		return n.length > t ? n.slice(0, t) + "…" : n;
	}
	function a(e) {
		try {
			return JSON.stringify(JSON.parse(e), null, 2);
		} catch {
			return e;
		}
	}
	let o = {
		operator: "operator",
		desk: "Desk user",
		controller: "controller prompt",
		message: "message"
	};
	var s = Tr(), c = nn(s), l = (e) => {
		var n = tc(), r = R(n), a = z(r, !0), s = B(r, 2), c = (e) => {
			var n = $s(), r = R(n), a = z(r, !0), o = z(B(r), !0);
			A(n), V((e) => {
				q(a, e), q(o, t.block.text);
			}, [() => i(t.block.text, 120)]), K(e, n);
		}, l = (e) => {
			var n = ec();
			Jr(n, () => qs(t.block.text), !0), A(n), K(e, n);
		};
		J(s, (e) => {
			t.block.source === "controller" ? e(c) : e(l, -1);
		}), A(n), V(() => {
			ni(n, "data-source", t.block.source), q(a, o[t.block.source] || t.block.source);
		}), K(e, n);
	}, u = (e) => {
		var n = nc();
		let r;
		Jr(n, () => qs(t.block.text), !0), A(n), V(() => r = Qr(n, 1, "assistant md svelte-5n3hzh", null, r, { live: t.block.live })), K(e, n);
	}, d = (e) => {
		let o = /* @__PURE__ */ N(() => Qs(n(), {
			title: t.block.name,
			target: t.block.target,
			intent: t.block.intent
		}));
		var s = oc(), c = R(s), l = R(c), u = z(l, !0), d = B(l, 2), f = z(d, !0), p = z(B(d, 2), !0);
		A(c);
		var m = B(c, 2), h = (e) => {
			var n = ac(), r = nn(n), i = (e) => {
				var n = rc(), r = z(n, !0);
				V((e) => q(r, e), [() => a(t.block.input)]), K(e, n);
			};
			J(r, (e) => {
				t.block.input && e(i);
			});
			var o = B(r, 2), s = (e) => {
				var n = ic(), r = z(n, !0);
				V(() => q(r, t.block.output)), K(e, n);
			};
			J(o, (e) => {
				t.block.output && e(s);
			}), K(e, n);
		};
		J(m, (e) => {
			W(r) && e(h);
		}), A(s), V((e) => {
			ni(s, "data-state", t.block.state), ni(s, "data-tool", t.block.id), ni(c, "aria-expanded", W(r)), q(u, W(o).label), q(f, e), q(p, t.block.state);
		}, [() => W(o).detail || i(t.block.input, 80)]), hr("click", c, () => L(r, !W(r))), K(e, s);
	}, f = (e) => {
		var n = sc(), r = z(n, !0);
		V(() => q(r, t.block.text)), K(e, n);
	}, p = (e) => {
		var n = lc(), r = R(n), a = z(r, !0), o = B(r, 2), s = (e) => {
			var n = cc(), r = z(n, !0);
			V(() => q(r, t.block.text)), K(e, n);
		}, c = /* @__PURE__ */ N(() => t.block.text.length > 160 || t.block.text.includes("\n"));
		J(o, (e) => {
			W(c) && e(s);
		}), A(n), V((e) => q(a, e), [() => i(t.block.text, 160)]), K(e, n);
	};
	J(c, (e) => {
		t.block.kind === "user" ? e(l) : t.block.kind === "assistant" ? e(u, 1) : t.block.kind === "tool" ? e(d, 2) : t.block.kind === "error" ? e(f, 3) : e(p, -1);
	}), K(e, s), Ge();
}
gr(["click"]);
// kit/attention.js
function dc(e, t, n) {
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
async function fc(e, t) {
	if (!await e.confirm(t.ask)) return !1;
	try {
		return await t.run(), e.toast(t.label + ": done"), !0;
	} catch (n) {
		return e.toast(t.label + " failed: " + String(n.message || n)), !1;
	}
}
var pc = Symbol.for("jobshop.desk.attention.pending"), mc = "jobshop-attention-pending";
function hc() {
	let e = globalThis;
	return e[pc] ??= /* @__PURE__ */ new Map();
}
function gc(e, t) {
	hc().set(e, t), globalThis.dispatchEvent?.(new CustomEvent(mc, { detail: {
		key: e,
		label: t
	} }));
}
// kit/plugin.svelte.js
function _c(e, t) {
	let n = t.api.slot(e, t.slot, t.context, t.target);
	return {
		update(r) {
			n && r.slot === t.slot && JSON.stringify(r.target) === JSON.stringify(t.target) ? n.context(r.context) : (n?.destroy(), n = r.api.slot(e, r.slot, r.context, r.target)), t = r;
		},
		destroy() {
			n?.destroy();
		}
	};
}
// plugins/chat/Pane.svelte
var vc = /* @__PURE__ */ G("<span> </span>"), yc = /* @__PURE__ */ G("· <a class=\"mono svelte-wy7o3a\"> </a>", 1), bc = /* @__PURE__ */ G("<button class=\"older svelte-wy7o3a\" data-testid=\"load-older\"> </button>"), xc = /* @__PURE__ */ G("<p class=\"err svelte-wy7o3a\"> </p>"), Sc = /* @__PURE__ */ G("<p class=\"none svelte-wy7o3a\">No messages yet.</p>"), Cc = /* @__PURE__ */ G("<div class=\"pending svelte-wy7o3a\" data-testid=\"chat-pending\"><div class=\"ptext svelte-wy7o3a\"> </div><span class=\"pstate svelte-wy7o3a\"> </span></div>"), wc = /* @__PURE__ */ G("<p class=\"status svelte-wy7o3a\" data-testid=\"chat-working\"><span class=\"spin svelte-wy7o3a\"></span> </p>"), Tc = /* @__PURE__ */ G("<button type=\"button\" data-testid=\"resolve\" class=\"svelte-wy7o3a\"> </button>"), Ec = /* @__PURE__ */ G("<div class=\"item svelte-wy7o3a\" data-testid=\"chat-item\"><div class=\"itxt svelte-wy7o3a\"><span class=\"pill svelte-wy7o3a\"> </span> <span class=\"ititle\"> </span></div> <div class=\"iacts svelte-wy7o3a\"></div></div>"), Dc = /* @__PURE__ */ G("<a class=\"mono svelte-wy7o3a\" data-testid=\"chat-ro-ticket\"> </a>"), Oc = /* @__PURE__ */ G("<p class=\"ro svelte-wy7o3a\" data-testid=\"chat-readonly\">Read-only: a worker's turn is not driven from here.<!></p>"), kc = /* @__PURE__ */ G("<span class=\"notice svelte-wy7o3a\" data-testid=\"chat-notice\"> </span>"), Ac = /* @__PURE__ */ G("<button type=\"button\" class=\"stop svelte-wy7o3a\" title=\"Cancel the turn (Esc)\" data-testid=\"chat-stop\">Stop</button> <button type=\"button\" class=\"interrupt svelte-wy7o3a\" title=\"Cancel the turn and send now (Ctrl+Enter)\" data-testid=\"chat-interrupt\">Interrupt</button>", 1), jc = /* @__PURE__ */ G("<form class=\"composer svelte-wy7o3a\"><textarea rows=\"2\" aria-label=\"Message\" enterkeyhint=\"send\" data-testid=\"chat-input\" class=\"svelte-wy7o3a\"></textarea> <div class=\"actions svelte-wy7o3a\"><button type=\"button\" class=\"history svelte-wy7o3a\" aria-label=\"Previous sent message\" title=\"Previous sent message (Up arrow)\">↑</button> <button type=\"button\" class=\"history svelte-wy7o3a\" aria-label=\"Next sent message\" title=\"Next sent message (Down arrow)\">↓</button> <!> <span class=\"grow svelte-wy7o3a\"></span> <span class=\"plugged svelte-wy7o3a\" data-testid=\"chat-composer-actions\"></span> <!> <button type=\"submit\" class=\"send svelte-wy7o3a\" data-testid=\"chat-send\"> </button></div></form>"), Mc = /* @__PURE__ */ G("<p class=\"ro svelte-wy7o3a\" data-testid=\"chat-closed\"> </p>"), Nc = /* @__PURE__ */ G("<p class=\"ro svelte-wy7o3a\" data-testid=\"chat-readonly\"> </p>"), Pc = /* @__PURE__ */ G("<span class=\"model svelte-wy7o3a\" data-testid=\"chat-model\"> </span>"), Fc = /* @__PURE__ */ G("<span class=\"usage svelte-wy7o3a\"> </span>"), Ic = /* @__PURE__ */ G("<section class=\"pane svelte-wy7o3a\" data-testid=\"chat-pane\"><header class=\"head svelte-wy7o3a\"><div class=\"who svelte-wy7o3a\"><div class=\"name svelte-wy7o3a\"> </div> <div class=\"meta svelte-wy7o3a\"><span class=\"mono svelte-wy7o3a\"> </span> <!> <!></div></div> <span class=\"state svelte-wy7o3a\" data-testid=\"chat-status\"> </span></header> <div class=\"scroll svelte-wy7o3a\" data-testid=\"journal\"><!> <!> <!> <!> <!></div> <!> <!> <footer class=\"foot svelte-wy7o3a\" data-testid=\"chat-usage\"><!> <!> <span class=\"grow svelte-wy7o3a\"></span> <span class=\"mono fid svelte-wy7o3a\"> </span></footer></section>");
function Lc(e, t) {
	We(t, !0);
	let n = pi(t, "context", 19, () => ({})), r = /* @__PURE__ */ N(() => String(n().session || "")), i = /* @__PURE__ */ N(() => String(n().item || "")), a = /* @__PURE__ */ N(() => W(i).split(":")[0] || ""), o = /* @__PURE__ */ N(() => W(i).slice(W(a).length + 1)), s = /* @__PURE__ */ N(() => !!n().readonly), c = /* @__PURE__ */ I(null);
	hn(() => {
		if (!W(r)) return;
		let e = new vi(t.api, W(r));
		return L(c, e, !0), e.open(), () => e.close();
	});
	let l = /* @__PURE__ */ I(null), u = /* @__PURE__ */ I(""), d = /* @__PURE__ */ I(qt({
		write: !1,
		reason: ""
	}));
	hn(() => {
		if (!W(r)) return;
		let e = !1;
		return L(l, null), L(u, ""), L(d, {
			write: !1,
			reason: ""
		}, !0), t.api.get("api/sessions?since=720h").then((t) => {
			e || L(l, t.find((e) => e.id === W(r)) || null, !0);
		}).catch(() => {}), t.api.get("api/chat").then((t) => {
			if (e) return;
			let n = (t.sessions || []).find((e) => e.id === W(r));
			L(u, n?.alive ? n.key : "", !0);
		}).catch(() => {}), t.api.session.access(W(r)).then((t) => {
			e || L(d, {
				write: !!t.write,
				reason: t.reason || ""
			}, !0);
		}).catch(() => {}), () => {
			e = !0;
		};
	});
	let f = /* @__PURE__ */ I(null);
	hn(() => {
		if (L(f, null), !W(o) || [
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
			e || L(f, t.ticket, !0);
		}).catch(() => {}), () => {
			e = !0;
		};
	});
	let p = /* @__PURE__ */ N(() => !!W(c)?.ended), m = /* @__PURE__ */ N(() => !!W(c)?.busy && !W(p)), h = /* @__PURE__ */ N(() => {
		let e = W(c)?.blocks ?? [];
		for (let t = e.length - 1; t >= 0; t--) {
			let n = e[t];
			if (n.kind === "tool" && (n.state === "running" || n.state === "input")) return Qs(W(l)?.agent, {
				title: n.name,
				target: n.target,
				intent: n.intent
			}).label;
		}
		return "";
	}), g = /* @__PURE__ */ N(() => W(d).write || !!W(u)), _ = /* @__PURE__ */ N(() => ({
		session: W(r),
		item: W(i),
		working: W(m)
	})), v = /* @__PURE__ */ N(() => W(u).startsWith("escalation:") ? "Escalation " + W(u).slice(11) : W(u).startsWith("attention:") ? "Resolve " + W(i) : W(u).startsWith("desk:") ? "Desk chat" : W(l)?.ticket?.title || "Session"), y = /* @__PURE__ */ I(""), b = /* @__PURE__ */ I(!1), x = /* @__PURE__ */ I(""), S = /* @__PURE__ */ I(qt([]));
	hn(() => {
		let e = (W(c)?.blocks ?? []).filter((e) => e.kind === "user").map((e) => e.text || "");
		W(S).length && W(S).some((t) => e.some((e) => e.includes(t.text))) && L(S, W(S).filter((t) => !e.some((e) => e.includes(t.text))), !0);
	});
	let C = /* @__PURE__ */ I(void 0), w = /* @__PURE__ */ I(!0), ee = /* @__PURE__ */ I(void 0), te = /* @__PURE__ */ I(-1), T = [], ne = "", re = null;
	function ie(e) {
		if (W(te) < 0) {
			if (!e) return;
			T = Xs(W(r), W(c)?.prompts() ?? []), ne = W(y);
		}
		let t = e ? Math.min(W(te) + 1, T.length - 1) : W(te) - 1;
		t < -1 || !T.length || (L(te, t, !0), L(y, t < 0 ? ne : T[t], !0), ir().then(() => {
			W(ee)?.focus(), W(ee)?.setSelectionRange(W(y).length, W(y).length), se();
		}));
	}
	async function ae() {
		W(c) && W(C) && (L(w, !1), re = {
			height: W(C).scrollHeight,
			top: W(C).scrollTop
		}, await W(c).loadOlder(), await ir(), re && W(C) && (W(C).scrollTop = re.top + W(C).scrollHeight - re.height, re = null));
	}
	function oe() {
		W(C) && L(w, W(C).scrollHeight - W(C).scrollTop - W(C).clientHeight < 80);
	}
	hn(() => {
		W(c)?.version, W(m), W(w) && !re && ir().then(() => {
			W(C) && (W(C).scrollTop = W(C).scrollHeight);
		});
	});
	function se() {
		W(ee) && (W(ee).style.height = "auto", W(ee).style.height = Math.min(W(ee).scrollHeight, Math.round(window.innerHeight * .4)) + "px");
	}
	async function ce(e) {
		let n = W(y).trim();
		if (n && !W(b)) {
			L(b, !0), L(x, "");
			try {
				await t.api.session.send(W(u) ? {
					key: W(u),
					text: n,
					interrupt: e
				} : {
					session: W(r),
					text: n,
					interrupt: e
				}), L(S, [...W(S), {
					text: n,
					state: W(m) && !e ? "queued for after this turn" : "sent, waiting for the session"
				}], !0), W(m) && !e && L(x, "Queued for after this turn"), Zs(W(r), n), L(te, -1), L(y, ""), L(w, !0), await ir(), se();
			} catch (e) {
				L(x, String(e.message || e), !0);
			} finally {
				L(b, !1);
			}
		}
	}
	async function le() {
		L(x, "");
		try {
			(await t.api.session.cancel(W(r))).cancelled || L(x, "No turn running");
		} catch (e) {
			L(x, String(e.message || e), !0);
		}
	}
	function ue(e) {
		if (e.isComposing) return;
		let t = e.currentTarget;
		e.key === "ArrowUp" && !e.shiftKey && t.selectionStart === t.selectionEnd && (W(te) >= 0 || !W(y).slice(0, t.selectionStart).includes("\n")) ? (e.preventDefault(), ie(!0)) : e.key === "ArrowDown" && !e.shiftKey && W(te) >= 0 ? (e.preventDefault(), ie(!1)) : e.key === "Enter" && !e.shiftKey ? (e.preventDefault(), ce(e.ctrlKey || e.metaKey)) : e.key === "Escape" && W(m) && (e.preventDefault(), e.stopPropagation(), le());
	}
	let de = /* @__PURE__ */ N(() => dc(t.api, W(a), W(o))), fe = /* @__PURE__ */ I(!1);
	async function pe(e) {
		L(fe, !0);
		try {
			await fc(t.api, e) && (gc(W(a) + ":" + W(o), e.label), t.api.closeSession());
		} finally {
			L(fe, !1);
		}
	}
	var me = Ic(), he = R(me), E = R(he), ge = R(E), _e = z(ge, !0), ve = B(ge, 2), ye = R(ve), be = z(ye, !0), D = B(ye, 2), xe = (e) => {
		var t = vc(), n = z(t);
		V(() => q(n, `· ${W(l).ticket.pool ?? ""}`)), K(e, t);
	};
	J(D, (e) => {
		W(l)?.ticket?.pool && e(xe);
	});
	var O = B(D, 2), Se = (e) => {
		var n = yc(), r = B(nn(n)), i = z(r, !0);
		V(() => {
			ni(r, "href", "?ticket=" + W(l).current), q(i, W(l).current);
		}), hr("click", r, (e) => {
			e.preventDefault(), t.api.openTicket(W(l).current);
		}), K(e, n);
	};
	J(O, (e) => {
		W(l)?.current && e(Se);
	}), A(ve), A(E);
	var k = z(B(E, 2), !0);
	A(he);
	var j = B(he, 2), Ce = R(j), we = (e) => {
		var t = bc(), n = z(t, !0);
		V(() => {
			t.disabled = W(c).loadingOlder, q(n, W(c).loadingOlder ? "Loading history…" : `Load older messages (${W(c).earlier} earlier)`);
		}), hr("click", t, ae), K(e, t);
	};
	J(Ce, (e) => {
		W(c) && W(c).earlier > 0 && e(we);
	});
	var Te = B(Ce, 2), Ee = (e) => {
		var t = xc(), n = z(t, !0);
		V(() => q(n, W(c).error)), K(e, t);
	};
	J(Te, (e) => {
		W(c)?.error && e(Ee);
	});
	var De = B(Te, 2);
	Hr(De, 17, () => W(c)?.blocks ?? [], Rr, (e, t) => {
		{
			let n = /* @__PURE__ */ N(() => W(l)?.agent);
			uc(e, {
				get block() {
					return W(t);
				},
				get agent() {
					return W(n);
				}
			});
		}
	}, (e) => {
		K(e, Sc());
	});
	var Oe = B(De, 2);
	Hr(Oe, 17, () => W(S), Rr, (e, t) => {
		var n = Cc(), r = R(n), i = z(r, !0), a = z(B(r), !0);
		A(n), V(() => {
			q(i, W(t).text), q(a, W(t).state);
		}), K(e, n);
	});
	var ke = B(Oe, 2), Ae = (e) => {
		var t = wc(), n = B(R(t), 1, !0);
		A(t), V(() => q(n, W(h) || "Working")), K(e, t);
	};
	J(ke, (e) => {
		W(m) && e(Ae);
	}), A(j), ui(j, (e) => L(C, e), () => W(C));
	var je = B(j, 2), Me = (e) => {
		var t = Ec(), n = R(t), r = R(n), i = z(r, !0), s = z(B(r, 2), !0);
		A(n);
		var c = B(n, 2);
		Hr(c, 21, () => W(de), (e) => e.label, (e, t) => {
			var n = Tc(), r = z(n, !0);
			V(() => {
				n.disabled = W(fe), q(r, W(t).label);
			}), hr("click", n, () => pe(W(t))), K(e, n);
		}), A(c), A(t), V(() => {
			q(i, W(a)), q(s, W(f)?.title || W(o));
		}), K(e, t);
	};
	J(je, (e) => {
		W(i) && e(Me);
	});
	var Ne = B(je, 2), Pe = (e) => {
		var n = Oc(), r = B(R(n)), i = (e) => {
			var n = Dc(), r = z(n, !0);
			V(() => {
				ni(n, "href", "?ticket=" + W(l).current), q(r, W(l).current);
			}), hr("click", n, (e) => {
				e.preventDefault(), t.api.openTicket(W(l).current);
			}), K(e, n);
		};
		J(r, (e) => {
			W(l)?.current && e(i);
		}), A(n), K(e, n);
	}, Fe = (e) => {
		var n = jc(), r = R(n);
		tt(r), ui(r, (e) => L(ee, e), () => W(ee));
		var i = B(r, 2), a = R(i), o = B(a, 2), s = B(o, 2), c = (e) => {
			var t = kc(), n = z(t, !0);
			V(() => q(n, W(x))), K(e, t);
		};
		J(s, (e) => {
			W(x) && e(c);
		});
		var l = B(s, 4);
		Yr(l, (e, t) => _c?.(e, t), () => ({
			api: t.api,
			slot: "chat:composer-actions",
			context: W(_)
		}));
		var u = B(l, 2), d = (e) => {
			var t = Ac(), n = nn(t), r = B(n, 2);
			V((e) => r.disabled = e, [() => !W(y).trim() || W(b)]), hr("click", n, le), hr("click", r, () => ce(!0)), K(e, t);
		};
		J(u, (e) => {
			W(m) && e(d);
		});
		var f = B(u, 2), p = z(f, !0);
		A(i), A(n), V((e) => {
			ni(r, "placeholder", W(m) ? "Queued for after this turn" : "Message"), o.disabled = W(te) < 0, f.disabled = e, q(p, W(m) ? "Queue" : "Send");
		}, [() => !W(y).trim() || W(b)]), mr("submit", n, (e) => {
			e.preventDefault(), ce(!1);
		}), hr("input", r, () => {
			L(te, -1), se();
		}), hr("keydown", r, ue), oi(r, () => W(y), (e) => L(y, e)), hr("click", a, () => ie(!0)), hr("click", o, () => ie(!1)), K(e, n);
	}, Ie = (e) => {
		var t = Mc(), n = z(t);
		V(() => q(n, `Session closed${W(c)?.ended && W(c).ended !== "closed" ? " · " + W(c).ended : ""}`)), K(e, t);
	}, Le = (e) => {
		var t = Nc(), n = z(t);
		V(() => q(n, `Read-only: ${(W(d).reason || "this session does not accept input") ?? ""}`)), K(e, t);
	};
	J(Ne, (e) => {
		W(s) ? e(Pe) : W(g) && !W(p) ? e(Fe, 1) : W(p) ? e(Ie, 2) : e(Le, -1);
	});
	var Re = B(Ne, 2), ze = R(Re), Be = (e) => {
		var t = Pc(), n = z(t);
		V(() => q(n, `${W(l).model ?? ""}${W(l).effort ? " · " + W(l).effort : ""}`)), K(e, t);
	};
	J(ze, (e) => {
		W(l)?.model && e(Be);
	});
	var Ve = B(ze, 2), He = (e) => {
		var t = Fc(), n = z(t);
		V((e, t) => q(n, `${e ?? ""} / ${t ?? ""} tokens`), [() => Js(W(c).usage.used), () => Js(W(c).usage.size)]), K(e, t);
	};
	J(Ve, (e) => {
		W(c)?.usage && e(He);
	});
	var Ue = z(B(Ve, 4), !0);
	A(Re), A(me), V(() => {
		ni(me, "aria-label", W(v)), q(_e, W(v)), q(be, W(r)), q(k, W(c)?.status ?? "connecting"), q(Ue, W(r));
	}), mr("scroll", j, oe), K(e, me), Ge();
}
gr([
	"click",
	"input",
	"keydown"
]);
// plugins/chat/ChatList.svelte
var Rc = /* @__PURE__ */ G("<div class=\"err svelte-162crgi\"> </div>"), zc = /* @__PURE__ */ G("<a><span></span> <span class=\"name svelte-162crgi\"> </span> <span class=\"mono faint id svelte-162crgi\"> </span></a>"), Bc = /* @__PURE__ */ G("<div class=\"faint pad0 svelte-162crgi\">No open operator sessions.</div>"), Vc = /* @__PURE__ */ G("<div class=\"card empty\" data-testid=\"escalations-empty\"><div class=\"muted\">No escalation sessions.</div></div>"), Hc = /* @__PURE__ */ G("<tr data-testid=\"escalation-row\"><td class=\"c-title\"> <div class=\"faint mono\"> <!></div></td><td class=\"c-st\"><span> </span></td><td class=\"muted full\"> </td><td class=\"c-st\"><span> </span></td><td class=\"muted c-time\"> </td></tr>"), Uc = /* @__PURE__ */ G("<div class=\"card scroll\"><table class=\"stack\" data-testid=\"escalations\"><thead><tr><th>Ticket</th><th>Cause</th><th>Why</th><th>State</th><th>Age</th></tr></thead><tbody></tbody></table></div>"), Wc = /* @__PURE__ */ G("<header class=\"vh svelte-162crgi\"><h1>Chat</h1> <span class=\"sub svelte-162crgi\">pool <span class=\"mono\"> </span></span></header> <div class=\"list svelte-162crgi\"><form class=\"new svelte-162crgi\"><textarea rows=\"3\" placeholder=\"What should the operator look at?\" data-testid=\"chat-new-input\" class=\"svelte-162crgi\"></textarea> <div class=\"row svelte-162crgi\"><button type=\"submit\" class=\"primary\" data-testid=\"chat-new\">New session</button></div></form> <!> <div data-testid=\"chat-list\"></div> <div class=\"plugged svelte-162crgi\" data-testid=\"chat-panels\"></div> <h2 class=\"st\">Escalations</h2> <!></div>", 1);
function Gc(e, t) {
	We(t, !0);
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
	let l = /* @__PURE__ */ N(() => (W(n)?.sessions ?? []).filter((e) => !e.key.startsWith("escalation:"))), u = /* @__PURE__ */ N(() => (W(r) ?? []).filter((e) => e.open)), d = /* @__PURE__ */ N(() => (W(r) ?? []).filter((e) => !e.open));
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
	var _ = Wc(), v = nn(_), b = B(R(v), 2), x = z(B(R(b)), !0);
	A(b), A(v);
	var S = B(v, 2), C = R(S), w = R(C);
	tt(w);
	var ee = z(B(w, 2));
	A(C);
	var te = B(C, 2), T = (e) => {
		var t = Rc(), n = z(t, !0);
		V(() => q(n, W(i))), K(e, t);
	};
	J(te, (e) => {
		W(i) && e(T);
	});
	var ne = B(te, 2);
	Hr(ne, 21, () => W(l), (e) => e.id, (e, t) => {
		var n = zc();
		let r;
		var i = R(n);
		let a;
		var o = B(i, 2), c = z(o, !0), l = z(B(o, 2), !0);
		A(n), V((e) => {
			ni(n, "href", "?view=chat&session=" + W(t).id), r = Qr(n, 1, "item svelte-162crgi", null, r, { active: W(s) === W(t).id }), a = Qr(i, 1, "dot", null, a, { accent: W(t).alive }), q(c, e), q(l, W(t).id);
		}, [() => f(W(t).key)]), hr("click", n, (e) => {
			e.preventDefault(), h(W(t).id, W(t).key.startsWith("attention:") ? W(t).key.slice(10) : "");
		}), K(e, n);
	}, (e) => {
		K(e, Bc());
	}), A(ne);
	var re = B(ne, 2);
	Yr(re, (e, t) => _c?.(e, t), () => ({
		api: t.api,
		slot: "chat:panel",
		context: { session: W(s) }
	}));
	var ie = B(re, 4), ae = (e) => {
		K(e, Vc());
	}, oe = (e) => {
		var t = Uc(), n = R(t), r = B(R(n));
		Hr(r, 21, () => [...W(u), ...W(d)], (e) => e.ticket, (e, t) => {
			var n = Hc();
			let r;
			var i = R(n), a = R(i, !0), o = B(a), c = R(o, !0), l = B(c), u = (e) => {
				var n = wr();
				V(() => q(n, `· ${W(t).held ?? ""}`)), K(e, n);
			};
			J(l, (e) => {
				W(t).held && W(t).held !== W(t).ticket && e(u);
			}), A(o), A(i);
			var d = B(i), f = R(d), g = z(f, !0);
			A(d);
			var _ = B(d), v = z(_, !0), y = B(_), b = R(y), x = z(b, !0);
			A(y);
			var S = z(B(y), !0);
			A(n), V((e) => {
				r = Qr(n, 1, "click", null, r, { sel: W(s) === W(t).session }), q(a, W(t).title), q(c, W(t).ticket), Qr(f, 1, `pill ${W(t).open ? "warn" : ""}`), q(g, p[W(t).cause] ?? W(t).cause ?? "escalation"), q(v, W(t).why), Qr(b, 1, `pill ${W(t).open ? "accent" : "ok"}`), q(x, W(t).open ? "open" : "closed" + (W(t).resolution ? " · " + W(t).resolution : "")), q(S, e);
			}, [() => m(W(t).open ? W(t).since : W(t).closed)]), hr("click", n, () => h(W(t).session, "ticket:" + W(t).ticket)), K(e, n);
		}), A(r), A(n), A(t), K(e, t);
	};
	J(ie, (e) => {
		W(r) && W(u).length + W(d).length === 0 ? e(ae) : W(r) && e(oe, 1);
	}), A(S), V((e) => {
		q(x, W(n)?.pool ?? "…"), ee.disabled = e;
	}, [() => W(o) || !W(a).trim()]), mr("submit", C, (e) => {
		e.preventDefault(), g();
	}), oi(w, () => W(a), (e) => L(a, e)), K(e, _), Ge();
}
gr(["click"]);
// plugins/chat/DrawerChat.svelte
var Kc = /* @__PURE__ */ G("<div class=\"wrap svelte-1i9xi0d\" data-testid=\"escalation-chat\"><div class=\"bar svelte-1i9xi0d\"><a class=\"svelte-1i9xi0d\">open in side panel</a></div> <div class=\"frame svelte-1i9xi0d\"><!></div></div>");
function qc(e, t) {
	We(t, !0);
	let n = pi(t, "context", 19, () => ({})), r = /* @__PURE__ */ N(() => n().ticket || {}), i = (e) => typeof e?.meta == "string" ? JSON.parse(e.meta) : e?.meta || {}, a = /* @__PURE__ */ N(() => {
		if (W(r).kind === "session") return W(r).id;
		let e = i(W(r)).escalation || i(W(r)).gateway;
		return e?.session ? e.session : W(r).holder && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(W(r).holder) ? W(r).holder : "";
	});
	var o = Tr(), s = nn(o), c = (e) => {
		var n = Kc(), r = R(n), i = z(r), o = B(r, 2), s = R(o);
		{
			let e = /* @__PURE__ */ N(() => ({ session: W(a) }));
			Lc(s, {
				get api() {
					return t.api;
				},
				get context() {
					return W(e);
				}
			});
		}
		A(o), A(n), V(() => ni(i, "href", "?session=" + W(a))), hr("click", i, (e) => {
			e.preventDefault(), t.api.openSession(W(a));
		}), K(e, n);
	};
	J(s, (e) => {
		W(a) && e(c);
	}), K(e, o), Ge();
}
gr(["click"]);
// plugins/chat/index.svelte.js
var Jc = new URL("./style.css", import.meta.url).href;
function Yc(e) {
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
			if (!document.querySelector(`link[href="${Jc}"]`)) {
				let e = document.createElement("link");
				e.rel = "stylesheet", e.href = Jc, document.head.append(e);
			}
			this.style.display = "contents", this.app = Mr(e, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && Ir(this.app), this.app = null;
		}
	};
}
for (let [e, t] of [
	["desk-chat-pane", Lc],
	["desk-chat-view", Gc],
	["desk-chat-drawer", qc]
]) customElements.get(e) || customElements.define(e, Yc(t));
