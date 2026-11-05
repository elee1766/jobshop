// node_modules/svelte/src/constants.js
var e = {}, t = Symbol("uninitialized"), n = "http://www.w3.org/1999/xhtml", r = "http://www.w3.org/2000/svg", i = Array.isArray, a = Array.prototype.indexOf, o = Array.prototype.includes, s = Array.from, c = Object.defineProperty, l = Object.getOwnPropertyDescriptor, u = Object.getOwnPropertyDescriptors, d = Object.prototype, f = Array.prototype, p = Object.getPrototypeOf, m = Object.isExtensible;
function h(e) {
	return typeof e == "function";
}
var g = () => {};
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
function y(e, t, n = !1) {
	return e === void 0 ? n ? t() : t : e;
}
function b(e, t) {
	if (Array.isArray(e)) return e;
	if (t === void 0 || !(Symbol.iterator in e)) return Array.from(e);
	let n = [];
	for (let r of e) if (n.push(r), n.length === t) break;
	return n;
}
// node_modules/svelte/src/internal/client/constants.js
var x = 1 << 24, S = 1024, C = 2048, w = 4096, ee = 8192, te = 16384, ne = 32768, re = 1 << 25, ie = 65536, ae = 1 << 19, oe = 1 << 20, se = 1 << 25, ce = 1 << 21, le = 1 << 22, ue = 1 << 23, de = Symbol("$state"), fe = Symbol("component"), pe = Symbol("legacy props"), me = Symbol(""), he = Symbol("attributes"), ge = Symbol("class"), T = Symbol("style"), _e = Symbol("text"), ve = Symbol("form reset"), ye = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), be = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function xe() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function Se(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function Ce() {
	console.warn("https://svelte.dev/e/select_multiple_invalid_value");
}
function we() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var E = !1;
function D(e) {
	E = e;
}
var O;
function k(t) {
	if (t === null) throw Se(), e;
	return O = t;
}
function Te() {
	return k(/* @__PURE__ */ rn(O));
}
function A(t) {
	if (E) {
		if (/* @__PURE__ */ rn(O) !== null) throw Se(), e;
		O = t;
	}
}
function Ee(e = 1) {
	if (E) {
		for (var t = e, n = O; t--;) n = /* @__PURE__ */ rn(n);
		O = n;
	}
}
function De(e = !0) {
	for (var t = 0, n = O;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ rn(n);
		e && n.remove(), n = i;
	}
}
function Oe(t) {
	if (!t || t.nodeType !== 8) throw Se(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function ke(e) {
	return e === this.v;
}
function Ae(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function je(e) {
	return !Ae(e, this.v);
}
function Me(e) {
	throw Error("https://svelte.dev/e/lifecycle_outside_component");
}
// node_modules/svelte/src/internal/client/errors.js
function Ne() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Pe(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function Fe(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function Ie() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function Le(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Re() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function ze(e) {
	throw Error("https://svelte.dev/e/props_invalid_value");
}
function Be() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Ve() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function He() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Ue() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/shared/context.js
function We(e) {
	let t = e.p;
	for (; t !== null && t.c === null;) t = t.p;
	return t?.c ?? null;
}
function Ge(e, t) {
	return e === null && Me(t), e.c ??= new Map(We(e) || void 0);
}
// node_modules/svelte/src/internal/client/context.js
var j = null;
function Ke(e) {
	j = e;
}
function qe(e) {
	return Ge(j, "getContext").get(e);
}
function Je(e, t = !1, n) {
	j = {
		p: j,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: U,
		l: null
	};
}
function Ye(e) {
	var t = j, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) bn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, j = t.p, Xe(e);
}
function Xe(e = {}) {
	return c(e, fe, { value: !0 }), e;
}
function Ze() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var Qe = [];
function $e() {
	var e = Qe;
	Qe = [], _(e);
}
function et(e) {
	if (Qe.length === 0 && !Ct) {
		var t = Qe;
		queueMicrotask(() => {
			t === Qe && $e();
		});
	}
	Qe.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var tt = ~(C | w | S);
function M(e, t) {
	e.f = e.f & tt | t;
}
function nt(e) {
	e.f & 512 || e.deps === null ? M(e, S) : M(e, w);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function rt(e, t) {
	e.f & 2048 ? t.set(e, C) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, w), M(e, S);
}
// node_modules/svelte/src/internal/client/dom/elements/misc.js
function it(e, t) {
	if (t) {
		let t = document.body;
		e.autofocus = !0, et(() => {
			document.activeElement === t && e.focus();
		});
	}
}
var at = !1;
function ot() {
	at || (at = !0, document.addEventListener("reset", (e) => {
		Promise.resolve().then(() => {
			if (!e.defaultPrevented) for (let t of e.target.elements) t[ve]?.();
		});
	}, { capture: !0 }));
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function st(e) {
	var t = H, n = U;
	Un(null), Wn(null);
	try {
		return e();
	} finally {
		Un(t), Wn(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function ct(e, t, n, r) {
	let i = Ze() ? ft : mt;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = U, c = lt(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				pn(e, s);
			}
			ut();
		}
	}
	var d = dt();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ N(e))).then(u).catch((e) => pn(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), ut());
	}) : f();
}
function lt() {
	var e = U, t = H, n = j, r = F;
	return function(i = !0) {
		Wn(e), Un(t), Ke(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function ut(e = !0) {
	Wn(null), Un(null), Ke(null), e && F?.deactivate();
}
function dt() {
	var e = U, t = e.b, n = F, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function ft(e) {
	var n = 2 | C;
	return U !== null && (U.f |= ae), {
		ctx: j,
		deps: null,
		effects: null,
		equals: ke,
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
var pt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function N(e, n, r) {
	let i = U;
	i === null && Ne();
	var a = void 0, o = zt(t), s = !H, c = /* @__PURE__ */ new Set();
	return Cn(() => {
		var t = U, n = v();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== ye && n.reject(e);
			}).finally(ut);
		} catch (e) {
			n.reject(e), ut();
		}
		var r = F;
		if (s) {
			if (t.f & 32768) var l = dt();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(pt);
			else for (let e of c.values()) e.reject(pt);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== pt && (r.activate(), t ? (o.f |= ue, Wt(o, t)) : (o.f & 8388608 && (o.f ^= ue), Wt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), vn(() => {
		for (let e of c) e.reject(pt);
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
function P(e) {
	let t = /* @__PURE__ */ ft(e);
	return Kn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function mt(e) {
	let t = /* @__PURE__ */ ft(e);
	return t.equals = je, t;
}
function ht(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) V(t[n]);
	}
}
function gt(e) {
	var n, r = U, i = e.parent;
	if (!Bn && i !== null && e.v !== t && i.f & 24576) return xe(), e.v;
	Wn(i);
	try {
		ht(e), n = ir(e);
	} finally {
		Wn(r);
	}
	return n;
}
function _t(e) {
	var t = gt(e);
	!e.equals(t) && (e.wv = tr(), (!F?.is_fork || e.deps === null) && (F === null ? e.v = t : (F.capture(e, t, !0), xt?.capture(e, t, !0)), e.deps === null)) ? M(e, S) : Bn || (I === null ? nt(e) : (_n() || F?.is_fork) && I.set(e, t));
}
function vt(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && st(() => {
		t.ac.abort(ye), t.ac = null;
	}), t.fn !== null && (t.teardown = g), sr(t, 0), An(t));
}
function yt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && cr(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var bt = null, F = null, xt = null, I = null, St = null, Ct = !1, wt = !1, Tt = null, Et = null, Dt = 0, Ot = 1, kt = class e {
	id = Ot++;
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
		bt === null ? bt = this : (bt.next = this, this.prev = bt), bt = this;
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
			for (var r of n.d) M(r, C), t(r);
			for (r of n.m) M(r, w), t(r);
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
					t.f ^= S;
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
		for (var t = Tt = [], n = [], r = Et = []; this.#a.length > 0;) {
			Dt++ > 1e3 && (this.#b(), At());
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
		if (Tt = null, Et = null, this.#d()) {
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
			this.#e.clear(), xt = this, Mt(n), Mt(t), xt = null, this.#i?.resolve();
			var o = F;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Lt.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= S;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= S : i & 4 ? t.push(r) : nr(r) && (i & 16 && this.add_dirty_reaction(r, w), cr(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), M(i, C), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), F = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) rt(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), I?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		F = this;
	}
	deactivate() {
		F = null, I = null;
	}
	flush() {
		try {
			wt = !0, F = this, this.#m();
		} finally {
			Dt = 0, St = null, Tt = null, Et = null, wt = !1, F = null, I = null, Lt.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(pt);
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
		this.#u || (this.#u = !0, et(() => {
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
			!wt && et(() => {
				t.started || t.flush();
			});
		}
		return F;
	}
	apply() {
		I = null;
	}
	schedule(e) {
		St = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? bt = e : t.prev = e, this.linked = !1;
		}
	}
};
function At() {
	try {
		Re();
	} catch (e) {
		pn(e, St);
	}
}
var jt = null;
function Mt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && nr(r) && (jt = /* @__PURE__ */ new Set(), cr(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && Nn(r), jt?.size > 0)) {
				Lt.clear();
				for (let e of jt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) jt.has(n) && (jt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || cr(n);
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
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), M(e, S);
		for (var n = e.first; n !== null;) Pt(n, t), n = n.next;
	}
}
function Ft(e) {
	M(e, S);
	for (var t = e.first; t !== null;) Ft(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var It = /* @__PURE__ */ new Set(), Lt = /* @__PURE__ */ new Map(), Rt = !1;
function zt(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: ke,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Bt(e, t) {
	let n = zt(e, t);
	return Kn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Vt(e, t = !1, n = !0) {
	let r = zt(e);
	return t || (r.equals = je), r;
}
function L(e, t, n = !1) {
	return H !== null && (!Hn || H.f & 131072) && Ze() && H.f & 4325394 && (Gn === null || !Gn.has(e)) && He(), Wt(e, n ? Yt(t) : t, Et);
}
var Ht = null, Ut = 0;
function Wt(e, t, n = null) {
	if (!e.equals(t)) {
		Bn ? Lt.set(e, t) : Lt.has(e) || Lt.set(e, e.v);
		var r = kt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && gt(t), I === null && nt(t);
		}
		e.wv = tr(), Ht = null, Ut = 0, Jt(e, C, n), Ht = null, Ze() && U !== null && U.f & 1024 && !(U.f & 96) && (Yn === null ? Xn([e]) : Yn.push(e)), !r.is_fork && It.size > 0 && !Rt && Gt();
	}
	return t;
}
function Gt() {
	Rt = !1;
	for (let e of It) {
		e.f & 1024 && M(e, w);
		let t;
		try {
			t = nr(e);
		} catch {
			t = !0;
		}
		t && cr(e);
	}
	It.clear();
}
function Kt(e, t = 1) {
	var n = W(e), r = t === 1 ? n++ : n--;
	return L(e, n), r;
}
function qt(e) {
	L(e, e.v + 1);
}
function Jt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ze(), a = r.length;
		if (Ut += a, Ut > 1e5 && Ht === null && (Ht = /* @__PURE__ */ new Set()), Ht !== null) {
			if (Ht.has(e)) return;
			Ht.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== U) {
				var l = (c & C) === 0;
				if (l && M(s, t), c & 131072) It.add(s);
				else if (c & 2) {
					var u = s;
					I?.delete(u), Jt(u, w, n);
				} else if (l) {
					var d = s;
					c & 16 && jt !== null && jt.add(d), n === null ? Nt(d) : n.push(d);
				}
			}
		}
	}
}
function Yt(e) {
	if (typeof e != "object" || !e || de in e || fe in e) return e;
	let n = p(e);
	if (n !== d && n !== f) return e;
	var r = /* @__PURE__ */ new Map(), a = i(e), o = /* @__PURE__ */ Bt(0), s = null, c = $n, u = (e) => {
		if ($n === c) return e();
		var t = H, n = $n;
		Un(null), er(c);
		var r = e();
		return Un(t), er(n), r;
	};
	return a && r.set("length", /* @__PURE__ */ Bt(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Be();
			var i = r.get(t);
			return i === void 0 ? u(() => {
				var e = /* @__PURE__ */ Bt(n.value, s);
				return r.set(t, e), e;
			}) : L(i, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var i = r.get(n);
			if (i === void 0) {
				if (n in e) {
					let e = u(() => /* @__PURE__ */ Bt(t, s));
					r.set(n, e), qt(o);
				}
			} else L(i, t), qt(o);
			return !0;
		},
		get(n, i, a) {
			if (i === de) return e;
			var o = r.get(i), c = i in n;
			if (o === void 0 && (!c || l(n, i)?.writable) && (o = u(() => /* @__PURE__ */ Bt(Yt(c ? n[i] : t), s)), r.set(i, o)), o !== void 0) {
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
			if (n === de) return !0;
			var i = r.get(n), a = i !== void 0 && i.v !== t || Reflect.has(e, n);
			return (i !== void 0 || (U !== null || H !== null) && (!a || l(e, n)?.writable)) && (i === void 0 && (i = u(() => /* @__PURE__ */ Bt(a ? Yt(e[n]) : t, s)), r.set(n, i)), W(i) === t) ? !1 : a;
		},
		set(e, n, i, c) {
			var d = r.get(n), f = n in e;
			if (a && n === "length") for (var p = i; p < d.v; p += 1) {
				var m = r.get(p + "");
				m === void 0 ? p in e && (m = u(() => /* @__PURE__ */ Bt(t, s)), r.set(p + "", m)) : L(m, t);
			}
			if (d === void 0) (!f || l(e, n)?.writable) && (d = u(() => /* @__PURE__ */ Bt(void 0, s)), L(d, Yt(i)), r.set(n, d));
			else {
				f = d.v !== t;
				var h = u(() => Yt(i));
				L(d, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, i), !f) {
				if (a && typeof n == "string") {
					var _ = r.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && L(_, v + 1);
				}
				qt(o);
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
			Ve();
		}
	});
}
function Xt(e) {
	try {
		if (typeof e == "object" && e && de in e) return e[de];
	} catch {}
	return e;
}
function Zt(e, t) {
	return Object.is(Xt(e), Xt(t));
}
var Qt, $t, en, tn;
function nn() {
	if (Qt === void 0) {
		Qt = window, $t = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		en = l(t, "firstChild").get, tn = l(t, "nextSibling").get, m(e) && (e[ge] = void 0, e[he] = null, e[T] = void 0, e.__e = void 0), m(n) && (n[_e] = void 0);
	}
}
function R(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function z(e) {
	return en.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function rn(e) {
	return tn.call(e);
}
function an(e, t) {
	if (!E) return /* @__PURE__ */ z(e);
	var n = /* @__PURE__ */ z(O);
	if (n === null) n = O.appendChild(R());
	else if (t && n.nodeType !== 3) {
		var r = R();
		return n?.before(r), k(r), r;
	}
	return t && dn(n), k(n), n;
}
function on(e, t = !1) {
	if (!E) {
		var n = /* @__PURE__ */ z(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ rn(n) : n;
	}
	if (t) {
		if (O?.nodeType !== 3) {
			var r = R();
			return O?.before(r), k(r), r;
		}
		dn(O);
	}
	return O;
}
function sn(e, t = !1) {
	if (!E) return /* @__PURE__ */ z(e);
	var n = an(e, t);
	return A(e), n;
}
function B(e, t = 1, n = !1) {
	let r = E ? O : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ rn(r);
	if (!E) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = R();
			return r === null ? i?.after(a) : r.before(a), k(a), a;
		}
		dn(r);
	}
	return k(r), r;
}
function cn(e) {
	e.textContent = "";
}
function ln() {
	return !1;
}
function un(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function dn(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function fn(e) {
	var t = U;
	if (t === null) return H.f |= ue, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	pn(e, t);
}
function pn(t, n) {
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
function mn(e) {
	U === null && (H === null && Le(e), Ie()), Bn && Fe(e);
}
function hn(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function gn(e, t) {
	var n = U;
	n !== null && n.f & 8192 && (e |= ee);
	var r = {
		ctx: j,
		deps: null,
		nodes: null,
		f: e | C | 512,
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
	if (e & 4) Tt === null ? kt.ensure().schedule(r) : Tt.push(r);
	else if (t !== null) {
		try {
			cr(r);
		} catch (e) {
			throw V(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= ie));
	}
	if (i !== null && (i.parent = n, n !== null && hn(i, n), H !== null && H.f & 2 && !(e & 64))) {
		var a = H;
		(a.effects ??= []).push(i);
	}
	return r;
}
function _n() {
	return H !== null && !Hn;
}
function vn(e) {
	let t = gn(8, null);
	return M(t, S), t.teardown = e, t;
}
function yn(e) {
	mn("$effect");
	var t = U.f;
	if (!H && t & 32 && j !== null && !j.i) {
		var n = j;
		(n.e ??= []).push(e);
	} else return bn(e);
}
function bn(e) {
	return gn(4 | oe, e);
}
function xn(e) {
	kt.ensure();
	let t = gn(64 | ae, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? Pn(t, () => {
			V(t), n(void 0);
		}) : (V(t), n(void 0));
	});
}
function Sn(e) {
	return gn(4, e);
}
function Cn(e) {
	return gn(le | ae, e);
}
function wn(e, t = 0) {
	return gn(8 | t, e);
}
function Tn(e, t = [], n = [], r = []) {
	ct(r, t, n, (t) => {
		gn(8, () => {
			e(...t.map(W));
		});
	});
}
function En(e, t = 0) {
	return gn(16 | t, e);
}
function Dn(e, t = 0) {
	return gn(x | t, e);
}
function On(e) {
	return gn(32 | ae, e);
}
function kn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = Bn, r = H;
		Vn(!0), Un(null);
		try {
			t.call(null);
		} catch (t) {
			pn(t, e.parent);
		} finally {
			Vn(n), Un(r);
		}
	}
}
function An(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && st(() => {
			e.abort(ye);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : V(n, t), n = r;
	}
}
function jn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || V(t), t = n;
	}
}
function V(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (Mn(e.nodes.start, e.nodes.end), n = !0), e.f |= re, An(e, t && !n), sr(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	kn(e), e.f ^= re, e.f |= te;
	var i = e.parent;
	i !== null && i.first !== null && Nn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function Mn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ rn(e);
		e.remove(), e = n;
	}
}
function Nn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function Pn(e, t, n = !0) {
	var r = [];
	e.f |= 256, Fn(e, r, !0);
	var i = () => {
		n && V(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function Fn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= ee;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				Fn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function In(e) {
	e.f &= -257, Ln(e, !0);
}
function Ln(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= ee, e.f & 1024 || (M(e, C), kt.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			Ln(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function Rn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ rn(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var zn = null, Bn = !1;
function Vn(e) {
	Bn = e;
}
var H = null, Hn = !1;
function Un(e) {
	H = e;
}
var U = null;
function Wn(e) {
	U = e;
}
var Gn = null;
function Kn(e) {
	H !== null && (H.f & 2097152 || H.f & 2) && (Gn ??= /* @__PURE__ */ new Set()).add(e);
}
var qn = null, Jn = 0, Yn = null;
function Xn(e) {
	Yn = e;
}
var Zn = 1, Qn = 0, $n = Qn;
function er(e) {
	$n = e;
}
function tr() {
	return ++Zn;
}
function nr(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (nr(a) && _t(a), a.wv > e.wv) return !0;
		}
		t & 512 && I === null && M(e, S);
	}
	return !1;
}
function rr(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(Gn !== null && Gn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? rr(a, t, !1) : t === a && (n ? M(a, C) : a.f & 1024 && M(a, w), Nt(a));
	}
}
function ir(e) {
	var t = qn, n = Jn, r = Yn, i = H, a = Gn, o = j, s = Hn, c = $n, l = e.f;
	qn = null, Jn = 0, Yn = null, H = l & 96 ? null : e, Gn = null, Ke(e.ctx), Hn = !1, $n = ++Qn, e.ac !== null && (st(() => {
		e.ac.abort(ye);
	}), e.ac = null);
	try {
		e.f |= ce;
		var u = e.fn, d = u();
		e.f |= ne;
		var f = ar(e);
		if (Ze() && Yn !== null && !Hn && f !== null && !(e.f & 6146)) for (var p = 0; p < Yn.length; p++) rr(Yn[p], e);
		if (i !== null && i !== e) {
			if (Qn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Qn;
			if (t !== null) for (let e of t) e.rv = Qn;
			Yn !== null && (r === null ? r = Yn : r.push(...Yn));
		}
		return e.f & 8388608 && (e.f ^= ue), d;
	} catch (t) {
		return ar(e), fn(t);
	} finally {
		e.f ^= ce, qn = t, Jn = n, Yn = r, H = i, Gn = a, Ke(o), Hn = s, $n = c;
	}
}
function ar(e) {
	var t = e.deps, n = F?.is_fork;
	if (qn !== null) {
		var r;
		if (n || sr(e, Jn), t !== null && Jn > 0) for (t.length = Jn + qn.length, r = 0; r < qn.length; r++) t[Jn + r] = qn[r];
		else e.deps = t = qn;
		if (_n() && e.f & 512) for (r = Jn; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && Jn < t.length && (sr(e, Jn), t.length = Jn);
	return t;
}
function or(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var i = a.call(r, e);
		if (i !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[i] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (qn === null || !o.call(qn, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && nt(c), c.ac !== null && st(() => {
			c.ac.abort(ye), c.ac = null, M(c, C);
		}), vt(c), sr(c, 0);
	}
}
function sr(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) or(e, n[r]);
}
function cr(e) {
	var t = e.f;
	if (!(t & 16384)) {
		M(e, S);
		var n = U;
		U = e;
		try {
			t & 16777232 ? jn(e) : An(e), kn(e);
			var r = ir(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Zn;
		} finally {
			U = n;
		}
	}
}
function W(e) {
	var t = !!(e.f & 2);
	if (zn?.add(e), H !== null && !Hn && !(U !== null && U.f & 16384) && (Gn === null || !Gn.has(e))) {
		var n = H.deps;
		if (H.f & 2097152) e.rv < Qn && (e.rv = Qn, qn === null && n !== null && n[Jn] === e ? Jn++ : qn === null ? qn = [e] : qn.push(e));
		else {
			H.deps ??= [], o.call(H.deps, e) || H.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [H] : o.call(r, H) || r.push(H);
		}
	}
	if (Bn && Lt.has(e)) return Lt.get(e);
	if (t) {
		var i = e;
		if (Bn) {
			var a = i.v;
			return (!(i.f & 1024) && i.reactions !== null || ur(i)) && (a = gt(i)), Lt.set(i, a), a;
		}
		var s = !(i.f & 512) && !Hn && H !== null && !!(H.f & 512), c = (i.f & ne) === 0;
		nr(i) && (s && (i.f |= 512), _t(i)), s && !c && (yt(i), lr(i));
	}
	if (I?.has(e)) return I.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function lr(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : o.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (yt(n), lr(n));
	}
}
function ur(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Lt.has(t) || t.f & 2 && ur(t)) return !0;
	return !1;
}
function dr(e) {
	var t = Hn;
	try {
		return Hn = !0, e();
	} finally {
		Hn = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var fr = Symbol("events"), pr = /* @__PURE__ */ new Set(), mr = /* @__PURE__ */ new Set();
function hr(e, t, n, r = {}) {
	function i(e) {
		if (r.capture || br.call(t, e), !e.cancelBubble) return st(() => n?.call(this, e));
	}
	return e.startsWith("pointer") || e.startsWith("touch") || e === "wheel" ? (i.__removed = !1, et(() => {
		i.__removed || t.addEventListener(e, i, r);
	})) : t.addEventListener(e, i, r), i;
}
function gr(e, t, n) {
	(t[fr] ??= {})[e] = n;
}
function _r(e) {
	for (var t = 0; t < e.length; t++) pr.add(e[t]);
	for (var n of mr) n(e);
}
var vr = null, yr = !1;
function br(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	vr = e, yr || (yr = !0, setTimeout(() => {
		yr = !1, vr = null;
	}));
	var o = 0, s = vr === e && e[fr];
	if (s) {
		var l = i.indexOf(s);
		if (l !== -1 && (t === document || t === window)) {
			e[fr] = t;
			return;
		}
		var u = i.indexOf(t);
		if (u === -1) return;
		l <= u && (o = l);
	}
	if (a = i[o] || e.target, a !== t) {
		c(e, "currentTarget", {
			configurable: !0,
			get() {
				return a || n;
			}
		});
		var d = H, f = U;
		Un(null), Wn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[fr]?.[r];
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
			e[fr] = t, delete e.currentTarget, Un(d), Wn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var xr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function Sr(e) {
	return xr?.createHTML(e) ?? e;
}
function Cr(e) {
	var t = un("template");
	return t.innerHTML = Sr(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function wr(e, t) {
	var n = U;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function Tr(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (E) return wr(O, null), O;
		i === void 0 && (i = Cr(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ z(i)));
		var t = r || $t ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ z(t), s = t.lastChild;
			wr(o, s);
		} else wr(t, t);
		return t;
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Er(e, t, n = "svg") {
	var r = !e.startsWith("<!>"), i = !!(t & 1), a = `<${n}>${r ? e : "<!>" + e}</${n}>`, o;
	return () => {
		if (E) return wr(O, null), O;
		if (!o) {
			var e = /* @__PURE__ */ z(Cr(a));
			if (i) for (o = document.createDocumentFragment(); /* @__PURE__ */ z(e);) o.appendChild(/* @__PURE__ */ z(e));
			else o = /* @__PURE__ */ z(e);
		}
		var t = o.cloneNode(!0);
		if (i) {
			var n = /* @__PURE__ */ z(t), r = t.lastChild;
			wr(n, r);
		} else wr(t, t);
		return t;
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Dr(e, t) {
	return /* @__PURE__ */ Er(e, t, "svg");
}
function Or() {
	if (E) return wr(O, null), O;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = R();
	return e.append(t, n), wr(t, n), e;
}
function G(e, t) {
	if (E) {
		var n = U;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = O), Te();
	} else e !== null && e.before(t);
}
// node_modules/svelte/src/utils.js
function kr(e) {
	return e.endsWith("capture") && e !== "gotpointercapture" && e !== "lostpointercapture";
}
var Ar = [
	"beforeinput",
	"click",
	"change",
	"dblclick",
	"contextmenu",
	"focusin",
	"focusout",
	"input",
	"keydown",
	"keyup",
	"mousedown",
	"mousemove",
	"mouseout",
	"mouseover",
	"mouseup",
	"pointerdown",
	"pointermove",
	"pointerout",
	"pointerover",
	"pointerup",
	"touchend",
	"touchmove",
	"touchstart"
];
function jr(e) {
	return Ar.includes(e);
}
var Mr = /* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split("."), Nr = {
	formnovalidate: "formNoValidate",
	ismap: "isMap",
	nomodule: "noModule",
	playsinline: "playsInline",
	readonly: "readOnly",
	defaultvalue: "defaultValue",
	defaultchecked: "defaultChecked",
	srcobject: "srcObject",
	novalidate: "noValidate",
	allowfullscreen: "allowFullscreen",
	disablepictureinpicture: "disablePictureInPicture",
	disableremoteplayback: "disableRemotePlayback"
};
function Pr(e) {
	return e = e.toLowerCase(), Nr[e] ?? e;
}
[...Mr];
var Fr = ["touchstart", "touchmove"];
function Ir(e) {
	return Fr.includes(e);
}
var Lr = [
	"textarea",
	"script",
	"style",
	"title"
];
function Rr(e) {
	return Lr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function zr(e) {
	let t = 0, n = zt(0), r;
	return () => {
		_n() && (W(n), wn(() => (t === 0 && (r = dr(() => e(() => qt(n)))), t += 1, () => {
			et(() => {
				--t, t === 0 && (r?.(), r = void 0, qt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var Br = ie | ae;
function Vr(e, t, n, r) {
	new Hr(e, t, n, r);
}
var Hr = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = E ? O : null;
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
	#m = zr(() => (this.#p = zt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = U;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = U.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = En(() => {
			if (E) {
				let e = Oe(this.#t);
				Te();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, Br), E && (this.#e = O);
	}
	#h() {
		try {
			this.#a = On(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		et(r), t && (this.#s = On(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? we() : (t = !0, n && Ue(), this.#s !== null && Pn(this.#s, () => {
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
					pn(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (te | re)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = On(() => e(this.#e)), et(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = R(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return On(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							pn(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(F)) : this.#u === 0 && (this.#e.before(e), this.#c = null, Pn(this.#o, () => {
					this.#o = null;
				}), this.#x(F));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = On(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				Rn(this.#a, e);
				let t = this.#n.pending;
				this.#o = On(() => t(this.#e));
			} else this.#x(F);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		rt(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = U, n = H, r = j;
		Wn(this.#i), Un(this.#i), Ke(this.#i.ctx);
		try {
			return kt.ensure(), e();
		} finally {
			Wn(t), Un(n), Ke(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && Pn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, et(() => {
			this.#d = !1, this.#p && Wt(this.#p, this.#l);
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
		this.#a &&= (V(this.#a), null), this.#o &&= (V(this.#o), null), this.#s &&= (V(this.#s), null), E && (k(this.#t), Ee(), k(De()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return On(() => {
						var r = U;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return pn(e, this.#i.parent), null;
				}
			}));
		};
		et(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					pn(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => pn(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function Ur(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[_e] ??= e.nodeValue) && (e[_e] = n, e.nodeValue = `${n}`);
}
function Wr(e, t) {
	return Kr(e, t);
}
var Gr = /* @__PURE__ */ new Map();
function Kr(t, { target: n, anchor: r, props: i = {}, events: a, context: o, intro: c = !0, transformError: l }) {
	nn();
	var u = void 0, d = xn(() => {
		var c = r ?? n.appendChild(R());
		Vr(c, { pending: () => {} }, (n) => {
			Je({});
			var r = j;
			if (o && (r.c = o), a && (i.$$events = a), E && wr(n, null), u = t(n, i) || Xe(), E && (U.nodes.end = O, O === null || O.nodeType !== 8 || O.data !== "]")) throw Se(), e;
			Ye();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = Ir(r);
					for (let e of [n, document]) {
						var a = Gr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), Gr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, br, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(s(pr)), mr.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = Gr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, br), t.delete(e), t.size === 0 && Gr.delete(r)) : t.set(e, i);
			}
			mr.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return qr.set(u, d), u;
}
var qr = /* @__PURE__ */ new WeakMap();
function Jr(e, t) {
	let n = qr.get(e);
	return n ? (qr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var Yr = class {
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
			if (n) In(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (In(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (V(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						Rn(r, t), t.append(R()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else V(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), Pn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (V(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = F, r = ln();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = R();
				i.append(a), this.#n.set(e, {
					effect: On(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, On(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else E && (this.anchor = O), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/snippet.js
function Xr(e, t, ...n) {
	var r = new Yr(e);
	En(() => {
		let e = t() ?? null;
		r.ensure(e, e && ((t) => e(t, ...n)));
	}, ie);
}
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function Zr(e, t, n = !1) {
	var r;
	E && (r = O, Te());
	var i = new Yr(e), a = n ? ie : 0;
	function o(e, t) {
		if (E) {
			var n = Oe(r);
			if (e !== parseInt(n.substring(1))) {
				var a = De();
				k(a), i.anchor = a, D(!1), i.ensure(e, t), D(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	En(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function Qr(e, t) {
	return t;
}
function $r(e, t, n) {
	for (var r = [], i = t.length, a, o = t.length, c = 0; c < i; c++) {
		let n = t[c];
		Pn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					ei(e, s(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --o;
		}, !1);
	}
	if (o === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			cn(d), d.append(u), e.items.clear();
		}
		ei(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function ei(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= se, Rn(a, document.createDocumentFragment())) : V(t[i], n);
	}
}
var ti;
function ni(e, t, n, r, a, o = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = E ? k(/* @__PURE__ */ z(u)) : u.appendChild(R());
	}
	E && Te();
	var d = null, f = /* @__PURE__ */ mt(() => {
		var e = n();
		return i(e) ? e : e == null ? [] : s(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = W(f);
			_.fallback = d, ii(_, n, c, t, r), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= se, oi(d, null, c)) : In(d) : Pn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: En(() => {
			var e = W(f), i = e.length;
			let s = !1;
			E && Oe(c) === "[!" != (i === 0) && (c = De(), k(c), D(!1), s = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = F, v = ln(), y = 0; y < i; y += 1) {
				E && O.nodeType === 8 && O.data === "]" && (c = O, s = !0, D(!1));
				var b = e[y], x = r(b, y), S = m ? null : l.get(x);
				S ? (S.v && Wt(S.v, b), S.i && Wt(S.i, y), v && _.unskip_effect(S.e)) : (S = ai(l, m ? c : ti ??= R(), b, x, y, a, t, n), m || (S.e.f |= se), l.set(x, S)), u.add(x);
			}
			if (i === 0 && o && !d && (m ? d = On(() => o(c)) : (d = On(() => o(ti ??= R())), d.f |= se)), i > u.size && Pe("", "", ""), E && i > 0 && k(De()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			s && D(!0), W(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, E && (c = O);
}
function ri(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function ii(e, t, n, r, i) {
	var a = !!(r & 8), o = t.length, c = e.items, l = ri(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < o; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < o; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (In(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= se, _ === l) oi(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), si(e, d, _), si(e, _, y), oi(_, y, n), d = _, p = [], m = [], l = ri(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) oi(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					si(e, S.prev, C.next), si(e, d, S), si(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), oi(_, l, n), si(e, _.prev, _.next), si(e, _, d === null ? e.effect.first : d.next), si(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = ri(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = ri(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (ei(e, s(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var w = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || w.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && w.push(l), l = ri(l.next);
		var ee = w.length;
		if (ee > 0) {
			var te = r & 4 && o === 0 ? n : null;
			if (a) {
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.measure();
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.fix();
			}
			$r(e, w, te);
		}
	}
	a && et(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function ai(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? zt(n) : /* @__PURE__ */ Vt(n, !1, !1) : null, l = o & 2 ? zt(i) : null;
	return {
		v: c,
		i: l,
		e: On(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function oi(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ rn(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function si(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/client/dom/blocks/svelte-component.js
function ci(e, t, n) {
	var r;
	E && (r = O, Te());
	var i = new Yr(e);
	En(() => {
		var e = t() ?? null;
		if (E && Oe(r) === "[" != (e !== null)) {
			var a = De();
			k(a), i.anchor = a, D(!1), i.ensure(e, e && ((t) => n(t, e))), D(!0);
		} else i.ensure(e, e && ((t) => n(t, e)));
	}, ie);
}
// node_modules/svelte/src/internal/client/dom/blocks/svelte-element.js
function li(e, t, n, i, a, o) {
	let s = E;
	E && Te();
	var c = null;
	E && O.nodeType === 1 && (c = O, Te());
	var l = E ? O : e, u = new Yr(l, !1);
	En(() => {
		let e = t() || null;
		var o = a ? a() : n || e === "svg" ? r : void 0;
		if (e === null) u.ensure(null, null);
		else return u.ensure(e, (t) => {
			if (e) {
				var n = E;
				if (c = E ? c : un(e, o), wr(c, c), i) {
					var r = null;
					E && Rr(e) && c.append(r = document.createComment(""));
					var a = E ? /* @__PURE__ */ z(c) : c.appendChild(R());
					E && (a === null ? D(!1) : k(a)), i(c, a), r?.remove();
				}
				U.nodes.end = c, n || t.before(c);
			}
			E && k(t);
		}), () => {};
	}, ie), vn(() => {}), s && (D(!0), k(l));
}
// node_modules/svelte/src/internal/client/dom/elements/attachments.js
function ui(e, t) {
	var n = void 0, r;
	Dn(() => {
		n !== (n = t()) && (r &&= (V(r), null), n && (r = On(() => {
			Sn(() => n(e));
		})));
	});
}
// node_modules/clsx/dist/clsx.mjs
function di(e) {
	var t, n, r = "";
	if (typeof e == "string" || typeof e == "number") r += e;
	else if (typeof e == "object") {
		if (Array.isArray(e)) {
			var i = e.length;
			for (t = 0; t < i; t++) e[t] && (n = di(e[t])) && (r && (r += " "), r += n);
		} else for (n in e) e[n] && (r && (r += " "), r += n);
	}
	return r;
}
function fi() {
	for (var e, t, n = 0, r = "", i = arguments.length; n < i; n++) (e = arguments[n]) && (t = di(e)) && (r && (r += " "), r += t);
	return r;
}
// node_modules/svelte/src/internal/shared/attributes.js
function pi(e) {
	return typeof e == "object" ? fi(e) : e ?? "";
}
var mi = [..." 	\n\r\f\xA0\v﻿"];
function hi(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || mi.includes(r[o - 1])) && (s === r.length || mi.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
function gi(e, t = !1) {
	var n = t ? " !important;" : ";", r = "";
	for (var i of Object.keys(e)) {
		var a = e[i];
		a != null && a !== "" && (r += " " + i + ": " + a + n);
	}
	return r;
}
function _i(e) {
	return e[0] !== "-" || e[1] !== "-" ? e.toLowerCase() : e;
}
function vi(e, t) {
	if (t) {
		var n = "", r, i;
		if (Array.isArray(t) ? (r = t[0], i = t[1]) : r = t, e) {
			e = String(e).replaceAll(/\/\*.*?\*\//g, "").trim();
			var a = !1, o = 0, s = !1, c = [];
			r && c.push(...Object.keys(r).map(_i)), i && c.push(...Object.keys(i).map(_i));
			var l = 0, u = -1;
			let t = e.length;
			for (var d = 0; d < t; d++) {
				var f = e[d];
				if (s ? f === "/" && e[d - 1] === "*" && (s = !1) : a ? a === f && (a = !1) : f === "/" && e[d + 1] === "*" ? s = !0 : f === "\"" || f === "'" ? a = f : f === "(" ? o++ : f === ")" && o--, !s && a === !1 && o === 0) {
					if (f === ":" && u === -1) u = d;
					else if (f === ";" || d === t - 1) {
						if (u !== -1) {
							var p = _i(e.substring(l, u).trim());
							if (!c.includes(p)) {
								f !== ";" && d++;
								var m = e.substring(l, d).trim();
								n += " " + m + ";";
							}
						}
						l = d + 1, u = -1;
					}
				}
			}
		}
		return r && (n += gi(r)), i && (n += gi(i, !0)), n = n.trim(), n === "" ? null : n;
	}
	return e == null ? null : String(e);
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function yi(e, t, n, r, i, a) {
	var o = e[ge];
	if (E || o !== n || o === void 0) {
		var s = hi(n, r, a);
		E && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[ge] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/style.js
function bi(e, t = {}, n, r) {
	for (var i in n) {
		var a = n[i];
		t[i] !== a && (n[i] == null ? e.style.removeProperty(i) : e.style.setProperty(i, a, r));
	}
}
function xi(e, t, n, r) {
	var i = e[T];
	if (E || i !== t) {
		var a = vi(t, r);
		E && a === e.getAttribute("style") || (a == null ? e.removeAttribute("style") : e.style.cssText = a), e[T] = t;
	} else r && (Array.isArray(r) ? (bi(e, n?.[0], r[0]), bi(e, n?.[1], r[1], "important")) : bi(e, n, r));
	return r;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/select.js
function Si(e, t) {
	t ? e.hasAttribute("selected") || e.setAttribute("selected", "") : e.removeAttribute("selected");
}
function Ci(e, t) {
	var n = !("__defaultValue" in e);
	(n || e.__defaultValue !== t) && (e.__defaultValue = t, wi(e, !n || "__value" in e));
}
function wi(e, t) {
	var n = e.__defaultValue, r = e.multiple, a = r ? n ?? [] : null;
	if (!r || i(a)) {
		var o = e.selectedIndex, s = t && r ? new Set(e.selectedOptions) : null;
		for (var c of e.options) {
			var l = Di(c);
			Si(c, r ? a.includes(l) : Zt(l, n));
		}
		if (t) {
			if (s !== null) for (c of e.options) {
				var u = s.has(c);
				c.selected !== u && (c.selected = u);
			}
			else e.selectedIndex !== o && (e.selectedIndex = o);
		}
	}
}
function Ti(e, t, n = !1) {
	if (e.multiple) {
		if (t == null) return;
		if (!i(t)) return Ce();
		for (var r of e.options) r.selected = t.includes(Di(r));
	} else {
		for (r of e.options) if (Zt(Di(r), t)) {
			r.selected = !0;
			return;
		}
		n && t === void 0 || (e.selectedIndex = -1);
	}
}
function Ei(e) {
	var t = new MutationObserver((t) => {
		t.every(Oi) || ("__defaultValue" in e && wi(e, !1), "__value" in e && Ti(e, e.__value));
	});
	t.observe(e, {
		childList: !0,
		subtree: !0,
		attributes: !0,
		attributeFilter: ["value"]
	}), vn(() => {
		t.disconnect();
	});
}
function Di(e) {
	return "__value" in e ? e.__value : e.value;
}
function Oi(e) {
	if (e.target.closest("selectedcontent") !== null) return !0;
	if (e.type === "childList") {
		var t = [...e.addedNodes, ...e.removedNodes];
		return t.length > 0 && t.every((e) => e.nodeName === "SELECTEDCONTENT");
	}
	return !1;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var ki = Symbol("class"), Ai = Symbol("style"), ji = Symbol("is custom element"), Mi = Symbol("is html"), Ni = be ? "link" : "LINK", Pi = be ? "input" : "INPUT", Fi = be ? "option" : "OPTION", Ii = be ? "select" : "SELECT", Li = be ? "progress" : "PROGRESS";
function Ri(e) {
	if (E) {
		var t = !1, n = () => {
			if (!t) {
				if (t = !0, e.hasAttribute("value")) {
					var n = e.value;
					zi(e, "value", null), e.value = n;
				}
				if (e.hasAttribute("checked")) {
					var r = e.checked;
					zi(e, "checked", null), e.checked = r;
				}
			}
		};
		e[ve] = n, et(n), ot();
	}
}
function zi(e, t, n, r) {
	var i = Hi(e);
	if (!(E && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Ni))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[me] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Wi(e).has(t) ? e[t] = n : (!E || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Bi(e, n, r, i, a = !1, o = !1) {
	E && a && e.nodeName === Pi && ("defaultValue" in r || "defaultChecked" in r || Ri(e));
	var s = Hi(e), c = s[ji], l = !s[Mi];
	let u = E && c;
	u && D(!1);
	var d = n || {}, f = e.nodeName === Fi, p = e.nodeName === Ii;
	try {
		for (var m in n) !(m in r) && m[0] + m[1] !== "$$" && (r[m] = null);
		r.class ? r.class = pi(r.class) : (i || r[ki]) && (r.class = null), r[Ai] && (r.style ??= null);
		var h = Wi(e);
		if (e.nodeName === Pi && "type" in r && ("value" in r || "__value" in r)) {
			var g = r.type;
			(g !== d.type || g === void 0 && e.hasAttribute("type")) && (d.type = g, zi(e, "type", g, o));
		}
		for (let a in r) {
			let u = r[a];
			if (f && a === "value" && u == null) e.value = e.__value = "", d[a] = u;
			else if (a === "class") yi(e, e.namespaceURI === "http://www.w3.org/1999/xhtml", u, i, n?.[ki], r[ki]), d[a] = u, d[ki] = r[ki];
			else if (a === "style") xi(e, u, n?.[Ai], r[Ai]), d[a] = u, d[Ai] = r[Ai];
			else {
				var _ = d[a];
				if (u !== _ || u === void 0 && e.hasAttribute(a)) {
					d[a] = u;
					var v = a[0] + a[1];
					if (v !== "$$") {
						if (v === "on") {
							let t = {}, n = "$$" + a, r = a.slice(2);
							var y = jr(r);
							if (kr(r) && (r = r.slice(0, -7), t.capture = !0), !y && _) {
								if (u != null) continue;
								e.removeEventListener(r, d[n], t), d[n] = null;
							}
							if (y) gr(r, e, u), _r([r]);
							else if (u != null) {
								function i(e) {
									d[a].call(this, e);
								}
								d[n] = hr(r, e, i, t);
							}
						} else if (a === "style") zi(e, a, u);
						else if (a === "autofocus") it(e, !!u);
						else if (!c && (a === "__value" || a === "value" && u != null)) e.__value = u, (_ == null || e.value !== u || u === 0 && e.nodeName === Li) && (e.value = u);
						else if (a === "selected" && f) Si(e, u);
						else {
							var b = a;
							l || (b = Pr(b));
							var x = b === "defaultValue" || b === "defaultChecked";
							if (p && b === "defaultValue") continue;
							if (u == null && !c && !x) {
								if (s[a] = null, b === "value" || b === "checked") {
									let t = e, r = n === void 0;
									if (b === "value") {
										let e = t.defaultValue;
										t.removeAttribute(b), t.defaultValue = e, t.value = t.__value = r ? e : null;
									} else {
										let e = t.defaultChecked;
										t.removeAttribute(b), t.defaultChecked = e, t.checked = r ? e : !1;
									}
								} else e.removeAttribute(a);
							} else x || (c || typeof u != "string") && h.has(b) ? (e[b] = u, b in s && (s[b] = t)) : typeof u != "function" && zi(e, b, u, o);
						}
					}
				}
			}
		}
	} finally {
		u && D(!0);
	}
	return d;
}
function Vi(e, t, n = [], r = [], i = [], a, o = !1, s = !1) {
	ct(i, n, r, (n) => {
		var r = void 0, i = {}, c = e.nodeName === Ii, l = !1;
		if (Dn(() => {
			var u = t(...n.map(W)), d = Bi(e, r, u, a, o, s);
			if (l && c) {
				var f = e;
				"defaultValue" in u && Ci(f, u.defaultValue), "value" in u && Ti(f, u.value);
			}
			for (let e of Object.getOwnPropertySymbols(i)) u[e] || V(i[e]);
			for (let t of Object.getOwnPropertySymbols(u)) {
				var p = u[t];
				t.description !== "@attach" || r && p === r[t] || (i[t] && V(i[t]), i[t] = On(() => ui(e, () => p))), d[t] = p;
			}
			r = d;
		}), c) {
			var u = e;
			Sn(() => {
				var e = r;
				"defaultValue" in e && Ci(u, e.defaultValue), Ti(u, e.value, !0), Ei(u);
			});
		}
		l = !0;
	});
}
function Hi(e) {
	return e[he] ??= {
		[ji]: e.nodeName.includes("-"),
		[Mi]: e.namespaceURI === n
	};
}
var Ui = /* @__PURE__ */ new Map();
function Wi(e) {
	var t = e.getAttribute("is") || e.nodeName, n = Ui.get(t);
	if (n) return n;
	Ui.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = u(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = p(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var Gi = !1;
function Ki(e) {
	var t = Gi;
	try {
		return Gi = !1, [e(), Gi];
	} finally {
		Gi = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
var qi = {
	get(e, t) {
		if (!e.exclude.has(t)) return e.props[t];
	},
	set(e, t) {
		return !1;
	},
	getOwnPropertyDescriptor(e, t) {
		if (!e.exclude.has(t) && t in e.props) return {
			enumerable: !0,
			configurable: !0,
			value: e.props[t]
		};
	},
	has(e, t) {
		return !e.exclude.has(t) && t in e.props;
	},
	ownKeys(e) {
		return Reflect.ownKeys(e.props).filter((t) => !e.exclude.has(t));
	}
};
/*#__NO_SIDE_EFFECTS__*/
function Ji(e, t, n) {
	return new Proxy({
		props: e,
		exclude: t
	}, qi);
}
var Yi = {
	get(e, t) {
		let n = e.props.length;
		for (; n--;) {
			let r = e.props[n];
			if (h(r) && (r = r()), typeof r == "object" && r && t in r) return r[t];
		}
	},
	set(e, t, n) {
		let r = e.props.length;
		for (; r--;) {
			let i = e.props[r];
			h(i) && (i = i());
			let a = l(i, t);
			if (a && a.set) return a.set(n), !0;
		}
		return !1;
	},
	getOwnPropertyDescriptor(e, t) {
		let n = e.props.length;
		for (; n--;) {
			let r = e.props[n];
			if (h(r) && (r = r()), typeof r == "object" && r && t in r) {
				let e = l(r, t);
				return e && !e.configurable && (e.configurable = !0), e;
			}
		}
	},
	has(e, t) {
		if (t === de || t === pe) return !1;
		for (let n of e.props) if (h(n) && (n = n()), n != null && t in n) return !0;
		return !1;
	},
	ownKeys(e) {
		let t = [];
		for (let n of e.props) if (h(n) && (n = n()), n) {
			for (let e in n) t.includes(e) || t.push(e);
			for (let e of Object.getOwnPropertySymbols(n)) t.includes(e) || t.push(e);
		}
		return t;
	}
};
function Xi(...e) {
	return new Proxy({ props: e }, Yi);
}
function Zi(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, c = !0, u = void 0, d = () => o && i ? (u ??= /* @__PURE__ */ ft(r), W(u)) : (c && (c = !1, s = o ? dr(r) : r), s);
	let f;
	if (a) {
		var p = de in e || pe in e;
		f = l(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = Ki(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && ze(t), f(m)));
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
	var v = !1, y = (n & 1 ? ft : mt)(() => (v = !1, g()));
	a && W(y);
	var b = U;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? W(y) : i && a ? Yt(e) : e;
			return L(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return Bn && v || b.f & 16384 ? y.v : W(y);
	});
}
// kit/plugin.svelte.js
function Qi(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href;
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ Bt(Yt({
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
			this.props.context = e ?? {};
		}
		get context() {
			return this.props.context;
		}
		set config(e) {
			this.props.config = e ?? {};
		}
		get config() {
			return this.props.config;
		}
		connectedCallback() {
			if (!document.querySelector(`link[href="${n}"]`)) {
				let e = document.createElement("link");
				e.rel = "stylesheet", e.href = n, document.head.append(e);
			}
			this.style.display = "contents", this.app ||= Wr(t, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && Jr(this.app), this.app = null;
		}
	});
}
function $i(e, t) {
	let n = Yt({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			Kt(r);
		}
	}), r = /* @__PURE__ */ Bt(0);
	yn(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, Kt(r);
			}, 1e3);
		});
		return () => {
			i(), clearTimeout(n);
		};
	});
	let i = !1, a = !1, o = "", s = (e, t) => {
		i = !0, a = !1;
		let r = (r) => {
			i = !1, t === o && (Object.assign(n, r), a && s(e, t));
		};
		e.get(t).then((e) => r({
			data: e,
			error: "",
			loading: !1
		}), (e) => r({
			error: String(e?.message || e),
			loading: !1
		}));
	};
	return yn(() => {
		let n = e(), c = t();
		W(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
function ea(e) {
	let t = Yt({});
	return yn(() => {
		let n = e();
		if (!n) return;
		let r = (e) => {
			for (let n of Object.keys(t)) n in e || delete t[n];
			Object.assign(t, e);
		};
		return r(n.params()), n.onParams(r);
	}), t;
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// node_modules/marked/lib/marked.esm.js
function ta() {
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
var na = ta();
function ra(e) {
	na = e;
}
var ia = { exec: () => null };
function aa(e) {
	let t = [];
	return (n) => {
		let r = Math.max(0, Math.min(3, n - 1)), i = t[r];
		return i || (i = e(r), t[r] = i), i;
	};
}
function K(e, t = "") {
	let n = typeof e == "string" ? e : e.source, r = {
		replace: (e, t) => {
			let i = typeof t == "string" ? t : t.source;
			return i = i.replace(q.caret, "$1"), n = n.replace(e, i), r;
		},
		getRegex: () => new RegExp(n, t)
	};
	return r;
}
var oa = ((e = "") => {
	try {
		return !!RegExp("(?<=1)(?<!1)" + e);
	} catch {
		return !1;
	}
})(), q = {
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
	nextBulletRegex: aa((e) => RegExp(`^ {0,${e}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`)),
	hrRegex: aa((e) => RegExp(`^ {0,${e}}((?:-[ 	]*){3,}|(?:_[ 	]*){3,}|(?:\\*[ 	]*){3,})(?:\\n+|$)`)),
	fencesBeginRegex: aa((e) => RegExp(`^ {0,${e}}(?:\`\`\`|~~~)`)),
	headingBeginRegex: aa((e) => RegExp(`^ {0,${e}}#{1,6}(?:\\s|$)`)),
	pedanticHeadingBeginRegex: aa((e) => RegExp(`^ {0,${e}}#`)),
	htmlBeginRegex: aa((e) => RegExp(`^ {0,${e}}(?:</?(?:${xa})(?: +|$|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)`, "i")),
	blockquoteBeginRegex: aa((e) => RegExp(`^ {0,${e}}>`))
}, sa = /^(?:[ \t]*(?:\n|$))+/, ca = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, la = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}(?:(?=`)\1`*|(?=~)\1~*) *(?=\n|$)|$)/, ua = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, da = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, fa = / {0,3}(?:[*+-]|\d{1,9}[.)])/, pa = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |fences|blockquote|heading|hr|html|table))+?)\n {0,3}(=+|-+)[ \t]*(?:\n+|$)/, ma = K(pa).replace(/bull/g, fa).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), ha = K(pa).replace(/bull/g, fa).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), ga = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table|[ \t]+\n)[^\n]+)*)/, _a = /^[^\n]+/, va = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\\n]|\n(?![ \t]*\n))+/, ya = K(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", va).replace("title", /(?:"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\\n]|\n(?!\n))*'|\((?:\\[\s\S]|[^()\\])*\))/).getRegex(), ba = K(/^(bull)([ \t][^\n]*?)?(?:\n|$)/).replace(/bull/g, fa).getRegex(), xa = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", Sa = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, Ca = K("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n*|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>[^\\n]*\\n*|$)|<![A-Z][\\s\\S]*?(?:>[^\\n]*\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>[^\\n]*\\n*|$)|</?(tag)(?:[ \\t]+|\\n|/?>|$)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!(?:script|pre|style|textarea)(?![a-z0-9-]))([a-z][a-z0-9-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</[a-z][a-z0-9-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", Sa).replace("tag", xa).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), wa = (e) => K(ga).replace("hr", ua).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", e).replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", xa).getRegex(), Ta = wa(/ {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]/), Ea = wa(/ {0,3}(?:[*+-]|\d{1,9}[.)])(?:[ \t]|\n|$)/), Da = {
	blockquote: K(/^( {0,3}> ?(?:(?![ \t]*(?:\n|$))(paragraph)|[^\n]*)(?:\n|$))+/).replace("paragraph", Ea).getRegex(),
	code: ca,
	def: ya,
	fences: la,
	heading: da,
	hr: ua,
	html: Ca,
	lheading: ma,
	list: ba,
	newline: sa,
	paragraph: Ta,
	table: ia,
	text: _a
}, Oa = K("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", ua).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", xa).getRegex(), ka = {
	...Da,
	lheading: ha,
	table: Oa,
	paragraph: K(ga).replace("hr", ua).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", Oa).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]+[^ \\t\\n]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", xa).getRegex()
}, Aa = {
	...Da,
	html: K("^ *(?:comment(?: *\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1>(?: *\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?>(?: *\\n{2,}|\\s*$))").replace("comment", Sa).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: ia,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: K(ga).replace("hr", ua).replace("heading", " *#{1,6} *[^\n]").replace("lheading", ma).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, ja = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, Ma = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, Na = /^( {2,}|\\)\n(?!\s*$)[ \t]*/, Pa = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, Fa = /[\p{P}\p{S}]/u, Ia = /[\s\p{P}\p{S}]/u, La = /[^\s\p{P}\p{S}]/u, Ra = K(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, Ia).getRegex(), za = /[\p{Pi}\p{Ps}"']/u, Ba = /(?!~)[\p{P}\p{S}]/u, Va = /(?!~)[\s\p{P}\p{S}]/u, Ha = /(?:[^\s\p{P}\p{S}]|~)/u, Ua = K(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", oa ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), Wa = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, Ga = K(Wa, "u").replace(/punct/g, Fa).getRegex(), Ka = K(Wa, "u").replace(/punct/g, Ba).getRegex(), qa = K(/^(?:\*+(?:((?!\*)(?!openQuote)punct)|([^\s*]))?)|^_+(?:((?!_)(?!openQuote)punct)|([^\s_]))?/, "u").replace(/openQuote/g, za).replace(/punct/g, Fa).getRegex(), Ja = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", Ya = K(Ja, "gu").replace(/notPunctSpace/g, La).replace(/punctSpace/g, Ia).replace(/punct/g, Fa).getRegex(), Xa = K(Ja, "gu").replace(/notPunctSpace/g, Ha).replace(/punctSpace/g, Va).replace(/punct/g, Ba).getRegex(), Za = K("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)[\\s](\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|(?:(?!\\*)punct|notPunctSpace)(\\*+)(?!\\*)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, La).replace(/punctSpace/g, Ia).replace(/punct/g, Fa).getRegex(), Qa = K("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, La).replace(/punctSpace/g, Ia).replace(/punct/g, Fa).getRegex(), $a = K("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)[\\s](_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)|(?:(?!_)punct|notPunctSpace)(_+)(?!_)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, La).replace(/punctSpace/g, Ia).replace(/punct/g, Fa).getRegex(), eo = K(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, Fa).getRegex(), to = K("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, La).replace(/punctSpace/g, Ia).replace(/punct/g, Fa).getRegex(), no = K(/\\(punct)/, "gu").replace(/punct/g, Fa).getRegex(), ro = K(/(`+)(?:[^`]|[^`][\s\S]*?[^`])\1(?!`)|`+|\\(punct)/, "gu").replace(/punct/g, Fa).getRegex(), io = K(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), ao = K(Sa).replace("(?:-->|$)", "-->").getRegex(), oo = K("^comment|^</[a-zA-Z][a-zA-Z0-9-]*\\s*>|^<[a-zA-Z][a-zA-Z0-9-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", ao).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), so = /\[(?:\\[\s\S]|[^\[\]\\])*\]/, co = K(/(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/).replace("brackets", so).getRegex(), lo = K(/^!?\[(label)\]\([ \t\n]*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?[ \t\n]*\)/).replace("label", co).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]+|(?=\))/).replace("title", /"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|\((?:\\[\s\S]|[^)\\])*\)/).getRegex(), uo = K(/^!?\[(label)\]\[(ref)\]/).replace("label", co).replace("ref", va).getRegex(), fo = K(/^!?\[(ref)\](?:\[\])?/).replace("ref", va).getRegex(), po = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\]){1,999}/, mo = K(/(?:[^\[\]\\`]*(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\]))){0,999}?[^\[\]\\`]*?/).replace("brackets", so).getRegex(), ho = K("reflink|nolink(?!\\()", "g").replace("reflink", K(/^!?\[(label)\]\[(ref)\]/).replace("label", mo).replace("ref", po).getRegex()).replace("nolink", K(/^!?\[(ref)\](?:\[\])?/).replace("ref", po).getRegex()).getRegex(), go = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, _o = K(/(?:mailto:email|xmpp:email(?:\/[A-Za-z0-9@.]+)?)/).replace(/email/g, /[A-Za-z0-9._+-]+@[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(), vo = {
	_backpedal: ia,
	anyPunctuation: no,
	anyPunctuationOutsideCode: ro,
	autolink: io,
	blockSkip: Ua,
	br: Na,
	code: Ma,
	del: ia,
	delLDelim: ia,
	delRDelim: ia,
	emStrongLDelim: Ga,
	emStrongRDelimAst: Ya,
	emStrongRDelimUnd: Qa,
	escape: ja,
	link: lo,
	nolink: fo,
	punctuation: Ra,
	reflink: uo,
	reflinkSearch: ho,
	tag: oo,
	text: Pa,
	url: ia
}, yo = {
	...vo,
	emStrongLDelim: qa,
	emStrongRDelimAst: Za,
	emStrongRDelimUnd: $a,
	link: K(/^!?\[(label)\]\((.*?)\)/).replace("label", co).getRegex(),
	reflink: K(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", co).getRegex()
}, bo = {
	...vo,
	emStrongRDelimAst: Xa,
	emStrongLDelim: Ka,
	delLDelim: eo,
	delRDelim: to,
	url: K(/^emailProtocol|^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("emailProtocol", _o).replace("protocol", go).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: K(/^(?:[^a-zA-Z0-9](?=emailProtocol)|(`+|~+|[^`~])(?:(?=[`~])|(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|wwwStart|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9](?=emailProtocol)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))))/).replace("protocol", go).replace(/emailProtocol/g, /(?:mailto|xmpp):/).replace("wwwStart", oa ? "(?<![^\\s*_~(])www\\." : "www\\.").getRegex()
}, xo = {
	...bo,
	br: K(Na).replace("{2,}", "*").getRegex(),
	text: K(bo.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, So = {
	normal: Da,
	gfm: ka,
	pedantic: Aa
}, Co = {
	normal: vo,
	gfm: bo,
	breaks: xo,
	pedantic: yo
}, wo = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, To = (e) => wo[e];
function Eo(e, t) {
	if (t) {
		if (q.escapeTest.test(e)) return e.replace(q.escapeReplace, To);
	} else if (q.escapeTestNoEncode.test(e)) return e.replace(q.escapeReplaceNoEncode, To);
	return e;
}
function Do(e) {
	return q.numericCharacterReference.test(e) ? e.replace(q.numericCharacterReferenceOrAmpersand, (e, t, n) => {
		if (t === void 0 && n === void 0) return "&amp;";
		let r = t === void 0 ? Number.parseInt(n, 16) : Number.parseInt(t, 10);
		return r === 0 || r > 1114111 || r >= 55296 && r <= 57343 ? "�" : r === 38 ? "&amp;" : String.fromCodePoint(r);
	}) : e;
}
function Oo(e) {
	try {
		e = encodeURI(e).replace(q.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function ko(e, t) {
	let n = e.replace(q.findPipe, (e, t, n) => {
		let r = !1, i = t;
		for (; --i >= 0 && n[i] === "\\";) r = !r;
		return r ? "|" : " |";
	}).split(q.splitPipe), r = 0;
	if (n[0].trim() || n.shift(), n.length > 0 && !n.at(-1)?.trim() && n.pop(), t) {
		if (n.length > t) n.splice(t);
		else for (; n.length < t;) n.push("");
	}
	for (; r < n.length; r++) n[r] = n[r].trim().replace(q.slashPipe, "|");
	return n;
}
function Ao(e, t, n) {
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
function jo(e) {
	let t = e.split("\n"), n = t.length - 1;
	for (; n >= 0 && q.blankLine.test(t[n]);) n--;
	return t.length - n <= 2 ? e : t.slice(0, n + 1).join("\n");
}
function Mo(e) {
	return e.trim().replace(q.multipleSpaceGlobal, " ").toLowerCase().toUpperCase().toLowerCase();
}
function No(e, t) {
	if (e.indexOf(t[0]) === -1 && e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
var Po = /* @__PURE__ */ new WeakMap();
function Fo(e) {
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
function Io(e, t) {
	let n = Fo(t), r = Po.get(e);
	r ? r.push(n) : Po.set(e, [n]);
}
function Lo(e) {
	let t = Po.get(e);
	t && (t.pop(), t.length === 0 && Po.delete(e));
}
function Ro(e, t, n) {
	let r = Po.get(e)?.at(-1);
	return r && r.maskedSrc === t ? r.lastPossibleEmStrongCloserIndex[n] : Fo(t).lastPossibleEmStrongCloserIndex[n];
}
function zo(e, t) {
	let n = Po.get(e)?.at(-1);
	return n && n.maskedSrc === t ? n.lastPossibleDelCloserIndex : Fo(t).lastPossibleDelCloserIndex;
}
function Bo(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function Vo(e, t, n, r, i) {
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
function Ho(e, t, n) {
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
function Uo(e, t, n, r) {
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
var Wo = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || na;
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
			let e = this.options.pedantic ? t[0] : jo(t[0]);
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
			let e = t[0], n = Ho(e, t[3] || "", this.rules);
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
				let t = Ao(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceTabChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: Ao(t[0], "\n"),
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
			raw: Ao(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = Ao(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteMarkerWhitespace, (e, t, n) => t + Bo(n, t.length)).replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
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
				let u = t[2].split("\n", 1)[0], d = t[1].length, f = this.options.pedantic ? Bo(u, d) : u.replace(this.rules.other.leadingSpaceTab, (e) => Bo(e, d)), p = e.split("\n", 1)[0], m = !f.trim(), h = 0;
				if (this.options.pedantic ? (h = 2, a = f.trimStart()) : m ? h = d + 1 : (h = f.search(this.rules.other.nonSpaceChar), h = h > 4 ? 1 : h, a = f.slice(h), h += d), m && this.rules.other.blankLine.test(p) && (e = e.substring(p.length + 1), i = !0), !i) {
					let t = this.rules.other.nextBulletRegex(h), n = this.rules.other.hrRegex(h), r = this.rules.other.fencesBeginRegex(h), i = this.options.pedantic ? this.rules.other.pedanticHeadingBeginRegex(h) : this.rules.other.headingBeginRegex(h), o = this.rules.other.htmlBeginRegex(h), s = this.rules.other.blockquoteBeginRegex(h);
					for (; e;) {
						let c = e.split("\n", 1)[0], l;
						if (p = c, this.options.pedantic ? (p = p.replace(this.rules.other.listReplaceNesting, "  "), l = p) : l = p.replace(this.rules.other.leadingSpaceTab, (e) => Bo(e)), r.test(p) || i.test(p) || o.test(p) || s.test(p) || t.test(p) || n.test(p)) break;
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
			let e = jo(t[0]);
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
			if (!this.rules.other.startAngleBracket.test(t[2]) && No(t[2], "()") !== -1) return;
			let e = Mo(t[1]), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: Ao(t[0], "\n"),
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = ko(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: Ao(t[0], "\n"),
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
			for (let e of i) a.rows.push(ko(e, a.header.length).map((e, t) => ({
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
				raw: Ao(t[0], "\n"),
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
			if (!this.options.pedantic && Uo(e, t[1], n, this.rules)) return;
			let r = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(r)) {
				if (!this.rules.other.endAngleBracket.test(r)) return;
				let e = Ao(r.slice(0, -1), "\\");
				if ((r.length - e.length) % 2 == 0) return;
			} else {
				let e = No(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let i = Ao(t[0].slice(0, -1), "\\");
			if ((t[0].length - i.length) % 2 == 0) return;
			let a = t[2], o = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(a);
				e && (a = e[1], o = e[3]);
			} else o = t[3] ? t[3].slice(1, -1) : "";
			return a = a.trim(), this.rules.other.startAngleBracket.test(a) && (a = this.options.pedantic && !this.rules.other.endAngleBracket.test(r) ? a.slice(1) : a.slice(1, -1)), Vo(t, {
				href: a && a.replace(this.rules.inline.anyPunctuation, "$1"),
				title: o && o.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let r = n[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && Uo(e, n[1], r, this.rules)) return;
			let i = t[Mo(n[2] || n[1])];
			if (!i) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return Vo(n, i, n[0], this.lexer, this.rules);
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
			let _ = i === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd, v = Ro(this, t, i), y = t.length - e.length + n;
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
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim, c = zo(this, t), l = t.length - e.length + n;
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
				text: e ? t[0] : Do(t[0]),
				escaped: e
			};
		}
	}
}, Go = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || na, this.options.tokenizer = this.options.tokenizer || new Wo(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			linkEmitted: !1,
			linkParenPossible: !0,
			top: !0
		};
		let t = {
			other: q,
			block: So.normal,
			inline: Co.normal
		};
		this.options.pedantic ? (t.block = So.pedantic, t.inline = Co.pedantic) : this.options.gfm && (t.block = So.gfm, t.inline = this.options.breaks ? Co.breaks : Co.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: So,
			inline: Co
		};
	}
	static lex(t, n) {
		return new e(n).lex(t);
	}
	static lexInline(t, n) {
		return new e(n).inlineTokens(t);
	}
	lex(e) {
		e = e.replace(q.carriageReturn, "\n"), this.blockTokens(e, this.tokens);
		for (let e = 0; e < this.inlineQueue.length; e++) {
			let t = this.inlineQueue[e];
			this.state.inRawBlock = !1, this.inlineTokens(t.src, t.tokens);
		}
		return this.inlineQueue = [], this.tokens;
	}
	blockTokens(e, t = [], n = !1) {
		this.tokenizer.lexer = this, this.options.pedantic && (e = e.replace(q.tabCharGlobal, "    ").replace(q.spaceLine, ""));
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
			if (e.charAt(0) !== "!" && Object.hasOwn(this.tokens.links, Mo(e.slice(n + 1, -1))) && !(n > 1 && this.linkInText(e.slice(1, n - 1)))) return !0;
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
		let n = e.includes("*") || e.includes("_") || e.includes("~") || this.options.hooks?.emStrongMask || this.tokenizer.emStrong !== Wo.prototype.emStrong || this.tokenizer.del !== Wo.prototype.del ? this.#n(e) : e;
		Io(this.tokenizer, n);
		try {
			return this.#t(e, n, t);
		} finally {
			Lo(this.tokenizer);
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
				if (!Object.hasOwn(this.tokens.links, Mo(t.slice(r + 1, -1)))) return t;
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
}, Ko = class {
	options;
	parser;
	constructor(e) {
		this.options = e || na;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(q.notSpaceStart)?.[0], i = e ? e.replace(q.endingNewline, "") + "\n" : "";
		return r ? "<pre><code class=\"language-" + Eo(r) + "\">" + (n ? i : Eo(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : Eo(i, !0)) + "</code></pre>\n";
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
		return `<code>${Eo(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, text: n, tokens: r, autolink: i }) {
		let a = i ? Eo(n, !0) : this.parser.parseInline(r), o = Oo(e);
		if (o === null) return a;
		e = Eo(o, i);
		let s = "<a href=\"" + e + "\"";
		return t && (s += " title=\"" + Eo(t) + "\""), s += ">" + a + "</a>", s;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = Oo(e);
		if (i === null) return Eo(n);
		e = i;
		let a = `<img src="${Eo(e)}" alt="${Eo(n)}"`;
		return t && (a += ` title="${Eo(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : Eo(e.text);
	}
}, qo = class {
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
}, Jo = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || na, this.options.renderer = this.options.renderer || new Ko(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new qo(), this.textRenderer.parser = this;
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
}, Yo = class {
	options;
	block;
	constructor(e) {
		this.options = e || na;
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
		return e ? Go.lex : Go.lexInline;
	}
	provideParser(e = this.block) {
		return e ? Jo.parse : Jo.parseInline;
	}
};
function Xo(e, t) {
	if (Array.isArray(t)) for (let n = 0; n < t.length; n++) e.push(t[n]);
	else e.push(t);
}
var Zo = new class {
	defaults = ta();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = Jo;
	Renderer = Ko;
	TextRenderer = qo;
	Lexer = Go;
	Tokenizer = Wo;
	Hooks = Yo;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (Xo(n, t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) Xo(n, this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) Xo(n, this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				Xo(n, this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					Xo(n, this.walkTokens(i, t));
				}) : e.tokens && Xo(n, this.walkTokens(e.tokens, t));
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
				let t = this.defaults.renderer || new Ko(this.defaults);
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
				let t = this.defaults.tokenizer || new Wo(this.defaults);
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
				let t = this.defaults.hooks || new Yo();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = Yo.passThroughHooks.has(n) ? (e) => {
						if (t.options.async && Yo.passThroughHooksRespectAsync.has(n)) return (async () => {
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
					return Xo(n, r.call(this, e)), t && Xo(n, t.call(this, e)), n;
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
		return Go.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return Jo.parse(e, t ?? this.defaults);
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
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? Go.lex : Go.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? Jo.parse : Jo.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? Go.lex : Go.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? Jo.parse : Jo.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + Eo(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function J(e, t) {
	return Zo.parse(e, t);
}
J.options = J.setOptions = function(e) {
	return Zo.setOptions(e), J.defaults = Zo.defaults, ra(J.defaults), J;
}, J.getDefaults = ta, J.defaults = na;
function Qo(...e) {
	return Zo.use(...e), J.defaults = Zo.defaults, ra(J.defaults), J;
}
J.use = Qo, J.walkTokens = function(e, t) {
	return Zo.walkTokens(e, t);
}, J.parseInline = Zo.parseInline, J.Parser = Jo, J.parser = Jo.parse, J.Renderer = Ko, J.TextRenderer = qo, J.Lexer = Go, J.lexer = Go.lex, J.Tokenizer = Wo, J.Hooks = Yo, J.parse = J, J.options, J.setOptions, J.walkTokens, J.parseInline, Jo.parse, Go.lex;
// node_modules/dompurify/dist/purify.es.mjs
function $o(e, t) {
	this.v = e, this.k = t;
}
function es(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function ts(e) {
	if (Array.isArray(e)) return e;
}
function ns(e, t) {
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
function rs() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function is(e, t) {
	return ts(e) || ns(e, t) || as(e, t) || rs();
}
function as(e, t) {
	if (e) {
		if (typeof e == "string") return es(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? es(e, t) : void 0;
	}
}
function os(e) {
	var t, n;
	function r(t, n) {
		try {
			var a = e[t](n), o = a.value, s = o instanceof $o;
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
os.prototype[typeof Symbol == "function" && Symbol.asyncIterator || "@@asyncIterator"] = function() {
	return this;
}, os.prototype.next = function(e) {
	return this._invoke("next", e);
}, os.prototype.throw = function(e) {
	return this._invoke("throw", e);
}, os.prototype.return = function(e) {
	return this._invoke("return", e);
};
var ss = Object.entries, cs = Object.setPrototypeOf, ls = Object.isFrozen, us = Object.getPrototypeOf, ds = Object.getOwnPropertyDescriptor, Y = Object.freeze, fs = Object.seal, ps = Object.create, ms = typeof Reflect < "u" && Reflect, hs = ms.apply, gs = ms.construct;
Y ||= function(e) {
	return e;
}, fs ||= function(e) {
	return e;
}, hs ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, gs ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var _s = Z(Array.prototype.forEach);
Array.prototype.indexOf;
var vs = Z(Array.prototype.lastIndexOf), ys = Z(Array.prototype.pop), bs = Z(Array.prototype.push), xs = Z(Array.prototype.slice), Ss = Z(Array.prototype.splice), Cs = Array.isArray, ws = Bs(Set), Ts = Z(Set.prototype.add), Es = Z(Set.prototype.has), Ds = Z(String.prototype.toLowerCase), Os = Z(String.prototype.toString), ks = Z(String.prototype.match), As = Z(String.prototype.replace), js = Z(String.prototype.indexOf), Ms = Z(String.prototype.trim), Ns = Z(Number.prototype.toString), Ps = Z(Boolean.prototype.toString), Fs = typeof BigInt > "u" ? null : Z(BigInt.prototype.toString), Is = typeof Symbol > "u" ? null : Z(Symbol.prototype.toString), Ls = Z(Object.prototype.hasOwnProperty), Rs = Z(Object.prototype.toString), X = Z(RegExp.prototype.test), zs = Bs(TypeError);
function Z(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return hs(e, t, n);
	};
}
function Bs(e) {
	return function() {
		return gs(e, [...arguments]);
	};
}
function Q(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : Ds;
	if (cs && cs(e, null), !Cs(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (ls(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function Vs(e) {
	for (let t = 0; t < e.length; t++) Ls(e, t) || (e[t] = null);
	return e;
}
function Hs(e) {
	let t = ps(null);
	for (let r of ss(e)) {
		var n = is(r, 2);
		let i = n[0], a = n[1];
		Ls(e, i) && (t[i] = Cs(a) ? Vs(a) : a && typeof a == "object" && a.constructor === Object ? Hs(a) : a);
	}
	return t;
}
function Us(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return Ns(e);
		case "boolean": return Ps(e);
		case "bigint": return Fs ? Fs(e) : "0";
		case "symbol": return Is ? Is(e) : "Symbol()";
		case "undefined": return Rs(e);
		case "function":
		case "object": {
			if (e === null) return Rs(e);
			let t = e, n = Ws(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : Rs(e);
			}
			return Rs(e);
		}
		default: return Rs(e);
	}
}
function Ws(e, t) {
	for (; e !== null;) {
		let n = ds(e, t);
		if (n) {
			if (n.get) return Z(n.get);
			if (typeof n.value == "function") return Z(n.value);
		}
		e = us(e);
	}
	function n() {
		return null;
	}
	return n;
}
function Gs(e, t) {
	for (; e !== null;) {
		let n = ds(e, t);
		if (n) return n.get ? Z(n.get) : null;
		e = us(e);
	}
	return null;
}
function Ks(e) {
	try {
		return X(e, ""), !0;
	} catch {
		return !1;
	}
}
var qs = Y(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), Js = Y(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), Ys = Y([
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
]), Xs = Y([
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
]), Zs = Y(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), Qs = Y([
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
]), $s = Y(["#text"]), ec = Y(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), tc = Y(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dominant-baseline.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.pointer-events.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-orientation.text-rendering.textlength.type.u1.u2.unicode.values.vector-effect.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), nc = Y(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), rc = Y([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), ic = fs(/{{[\w\W]*|^[\w\W]*}}/g), ac = fs(/<%[\w\W]*|^[\w\W]*%>/g), oc = fs(/\${[\w\W]*/g), sc = fs(/^data-[\-\w.\u00B7-\uFFFF]+$/), cc = fs(/^aria-[\-\w]+$/), lc = fs(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), uc = fs(/^(?:\w+script|data):/i), dc = fs(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), fc = fs(/^html$/i), pc = fs(/^[a-z][.\w]*(-[.\w]+)+$/i), mc = fs(/<[/\w!]/g), hc = fs(/<[/\w]/g), gc = fs(/<\/no(script|embed|frames)/i), _c = fs(/\/>/i), $ = {
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
}, vc = Y(Q({}, [
	"style",
	"script",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"plaintext",
	"noscript"
])), yc = function() {
	return typeof window > "u" ? null : window;
}, bc = function(e, t) {
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
}, xc = function() {
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
}, Sc = function(e, t, n, r) {
	return Ls(e, t) && Cs(e[t]) ? Q(r.base ? Hs(r.base) : {}, e[t], r.transform) : n;
}, Cc = function(e, t, n) {
	let r = Ls(e, t) ? e[t] : void 0;
	return r && typeof r == "object" ? Hs(r) : n();
};
function wc() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : yc(), t = (e) => wc(e);
	if (t.version = "3.4.17", t.removed = [], !e || !e.document || e.document.nodeType !== $.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = Ws(d, "cloneNode"), p = Ws(d, "remove"), m = Ws(d, "removeAttributeNode"), h = Ws(d, "nextSibling"), g = Ws(d, "childNodes"), _ = Ws(d, "parentNode"), v = Ws(d, "shadowRoot"), y = Ws(d, "attributes"), b = o && o.prototype ? Ws(o.prototype, "nodeType") : null, x = o && o.prototype ? Ws(o.prototype, "nodeName") : null, S = o && o.prototype ? Ws(o.prototype, "ownerDocument") : null, C = function(e) {
		return b ? b(e) : e.nodeType;
	}, w = function(e) {
		return x ? x(e) : e.nodeName;
	}, ee = Gs(d, "namespaceURI"), te = function(e) {
		return ee ? ee(e) : e.namespaceURI;
	};
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let ne, re = "", ie, ae = !1, oe = 0, se = function() {
		if (oe > 0) throw zs("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
	}, ce = function(e) {
		se(), oe++;
		try {
			return ne.createHTML(e);
		} finally {
			oe--;
		}
	}, le = function(e) {
		se(), oe++;
		try {
			return ne.createScriptURL(e);
		} finally {
			oe--;
		}
	}, ue = function() {
		return ae ||= (ie = bc(u, i), !0), ie;
	}, de = n, fe = de.implementation, pe = de.createNodeIterator, me = de.createDocumentFragment, he = de.getElementsByTagName, ge = r.importNode, T = xc();
	t.isSupported = typeof ss == "function" && typeof _ == "function" && fe && fe.createHTMLDocument !== void 0;
	let _e = ic, ve = ac, ye = oc, be = sc, xe = cc, Se = uc, Ce = dc, we = pc, E = lc, D = null, O = Q({}, [
		...qs,
		...Js,
		...Ys,
		...Zs,
		...$s
	]), k = null, Te = Q({}, [
		...ec,
		...tc,
		...nc,
		...rc
	]), A = Object.seal(ps(null, {
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
	})), Ee = null, De = null, Oe = Object.seal(ps(null, {
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
	})), ke = !0, Ae = !0, je = !1, Me = !0, Ne = !1, Pe = !0, Fe = !1, Ie = !1, Le = null, Re = null, ze = !1, Be = !1, Ve = !1, He = !1, Ue = !0, We = !1, Ge = "user-content-", j = !0, Ke = !1, qe = null, Je = {}, Ye = null, Xe = Q({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Ze = null, Qe = Q({}, [
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
	]), tt = "http://www.w3.org/1998/Math/MathML", M = "http://www.w3.org/2000/svg", nt = "http://www.w3.org/1999/xhtml", rt = nt, it = !1, at = null, ot = Q({}, [
		tt,
		M,
		nt
	], Os), st = Y([
		"mi",
		"mo",
		"mn",
		"ms",
		"mtext"
	]), ct = Q({}, st), lt = Y(["annotation-xml"]), ut = Q({}, lt), dt = Q({}, [
		"title",
		"style",
		"font",
		"a",
		"script"
	]), ft = null, pt = ["application/xhtml+xml", "text/html"], N = null, P = null, mt = n.createElement("form"), ht = function(e) {
		return e instanceof RegExp || e instanceof Function;
	}, gt = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		if (P && P === e) return;
		e && typeof e == "object" || (e = {}), e = Hs(e), ft = pt.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, N = ft === "application/xhtml+xml" ? Os : Ds, D = Sc(e, "ALLOWED_TAGS", O, { transform: N }), k = Sc(e, "ALLOWED_ATTR", Te, { transform: N }), at = Sc(e, "ALLOWED_NAMESPACES", ot, { transform: Os }), $e = Sc(e, "ADD_URI_SAFE_ATTR", et, {
			transform: N,
			base: et
		}), Ze = Sc(e, "ADD_DATA_URI_TAGS", Qe, {
			transform: N,
			base: Qe
		}), Ye = Sc(e, "FORBID_CONTENTS", Xe, { transform: N }), Ee = Sc(e, "FORBID_TAGS", Hs({}), { transform: N }), De = Sc(e, "FORBID_ATTR", Hs({}), { transform: N }), Je = Ls(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? Hs(e.USE_PROFILES) : e.USE_PROFILES : !1, ke = e.ALLOW_ARIA_ATTR !== !1, Ae = e.ALLOW_DATA_ATTR !== !1, je = e.ALLOW_UNKNOWN_PROTOCOLS || !1, Me = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, Ne = e.SAFE_FOR_TEMPLATES || !1, Pe = e.SAFE_FOR_XML !== !1, Fe = e.WHOLE_DOCUMENT || !1, Be = e.RETURN_DOM || !1, Ve = e.RETURN_DOM_FRAGMENT || !1, He = e.RETURN_TRUSTED_TYPE || !1, ze = e.FORCE_BODY || !1, Ue = e.SANITIZE_DOM !== !1, We = e.SANITIZE_NAMED_PROPS || !1, j = e.KEEP_CONTENT !== !1, Ke = e.IN_PLACE || !1, E = Ks(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : lc, rt = typeof e.NAMESPACE == "string" ? e.NAMESPACE : nt, ct = Cc(e, "MATHML_TEXT_INTEGRATION_POINTS", () => Q({}, st)), ut = Cc(e, "HTML_INTEGRATION_POINTS", () => Q({}, lt));
		let t = Cc(e, "CUSTOM_ELEMENT_HANDLING", () => ps(null));
		if (A = ps(null), Ls(t, "tagNameCheck") && ht(t.tagNameCheck) && (A.tagNameCheck = t.tagNameCheck), Ls(t, "attributeNameCheck") && ht(t.attributeNameCheck) && (A.attributeNameCheck = t.attributeNameCheck), Ls(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (A.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), fs(A), Ne && (Ae = !1), Ve && (Be = !0), Je && (D = Q({}, $s), k = ps(null), Je.html === !0 && (Q(D, qs), Q(k, ec)), Je.svg === !0 && (Q(D, Js), Q(k, tc), Q(k, rc)), Je.svgFilters === !0 && (Q(D, Ys), Q(k, tc), Q(k, rc)), Je.mathMl === !0 && (Q(D, Zs), Q(k, nc), Q(k, rc))), Oe.tagCheck = null, Oe.attributeCheck = null, Ls(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? Oe.tagCheck = e.ADD_TAGS : Cs(e.ADD_TAGS) && (D === O && (D = Hs(D)), Q(D, e.ADD_TAGS, N))), Ls(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? Oe.attributeCheck = e.ADD_ATTR : Cs(e.ADD_ATTR) && (k === Te && (k = Hs(k)), Q(k, e.ADD_ATTR, N))), Ls(e, "ADD_FORBID_CONTENTS") && Cs(e.ADD_FORBID_CONTENTS) && (Ye === Xe && (Ye = Hs(Ye)), Q(Ye, e.ADD_FORBID_CONTENTS, N)), j && (D["#text"] = !0), Fe && Q(D, [
			"html",
			"head",
			"body"
		]), D.table && (Q(D, ["tbody"]), delete Ee.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw zs("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw zs("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
			let t = ne;
			ne = e.TRUSTED_TYPES_POLICY;
			try {
				re = ce("");
			} catch (e) {
				throw ne = t, e;
			}
		} else e.TRUSTED_TYPES_POLICY === null ? (ne = void 0, re = "") : (ne === void 0 && (ne = ue()), ne && typeof re == "string" && (re = ce("")));
		Y && Y(e), P = e;
	}, _t = Q({}, [
		...Js,
		...Ys,
		...Xs
	]), vt = Q({}, [...Zs, ...Qs]), yt = function(e, t, n) {
		return t.namespaceURI === nt ? e === "svg" : t.namespaceURI === tt ? e === "svg" && (n === "annotation-xml" || ct[n]) : !!_t[e];
	}, bt = function(e, t, n) {
		return t.namespaceURI === nt ? e === "math" : t.namespaceURI === M ? e === "math" && ut[n] : !!vt[e];
	}, F = function(e, t, n) {
		return t.namespaceURI === M && !ut[n] || t.namespaceURI === tt && !ct[n] ? !1 : !vt[e] && (dt[e] || !_t[e]);
	}, xt = function(e) {
		let t = te(e), n = Ds(w(e)), r = _(e), i, a;
		return r && C(r) === $.element ? (i = { namespaceURI: te(r) }, a = Ds(w(r))) : (i = { namespaceURI: rt }, a = "template"), at[t] ? t === M ? yt(n, i, a) : t === tt ? bt(n, i, a) : t === nt ? F(n, i, a) : !!(ft === "application/xhtml+xml" && at[t]) : !1;
	}, I = function(e) {
		bs(t.removed, { element: e });
		try {
			_(e).removeChild(e);
		} catch {
			if (p(e), !_(e)) throw zs("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
		}
	}, St = function(e, t, n) {
		try {
			m(e, t);
		} catch {
			try {
				e.removeAttribute(n);
			} catch {}
		}
	}, Ct = function(e) {
		Dt(e);
		let t = g(e);
		t && _s(xs(t), (e) => {
			try {
				p(e);
			} catch {}
		});
		let n = C(e) === $.element ? y(e) : null;
		if (n) for (let t = n.length - 1; t >= 0; --t) {
			let r = n[t], i = r && r.name;
			typeof i == "string" && St(e, r, i);
		}
	}, wt = function(e, n, r) {
		if (!r) try {
			r = n.getAttributeNode(e);
		} catch {
			r = null;
		}
		bs(t.removed, {
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
				I(n);
			} catch {}
			else try {
				n.setAttribute(e, "");
			} catch {}
		}
	}, Tt = function(e) {
		let t = y(e);
		if (t) for (let n = t.length - 1; n >= 0; --n) {
			let r = t[n], i = r && r.name;
			typeof i != "string" || k[N(i)] || St(e, r, i);
		}
	}, Et = function(e, t) {
		let n = arguments.length > 2 && arguments[2] !== void 0 && arguments[2], r = [e];
		for (; r.length > 0;) {
			let e = r.pop();
			if (t(e) === !1) continue;
			if (n && C(e) === $.element) {
				let t = v(e);
				L(t) && r.push(t);
			}
			let i = g(e);
			if (i) for (let e = i.length - 1; e >= 0; --e) r.push(i[e]);
		}
	}, Dt = function(e) {
		Et(e, (e) => {
			C(e) === $.element && Tt(e);
		});
	}, Ot = function(e) {
		let t = [];
		return Et(e, (e) => {
			t.push(e);
		}, !0), t;
	}, kt = function(e, t) {
		let n = ws();
		_s(Ot(e), (e) => {
			C(e) === $.element && Ts(n, e);
		}), _s(t, (e) => {
			C(e) === $.element && !Es(n, e) && Tt(e);
		});
	}, At = function(e, n) {
		let r = !1;
		return _s(t.removed, (t) => {
			t.element && (r ||= t.element === e, Dt(t.element));
		}), kt(e, n), r;
	}, jt = function(e, t) {
		let n = qe, r = null, i = () => (r || (r = ws(), _s(t, (e) => {
			Ts(r, e);
		})), r);
		for (let t = 0;; t++) {
			let r = Ot(e), a = [e], o = !1;
			if (_s(r, (e) => {
				C(e) === $.documentFragment ? bs(a, e) : !o && !Es(n, e) && (o = Es(i(), e));
			}), !o) return;
			if (t >= 8) throw zs("the tree kept changing during in-place sanitization; refusing to return it");
			_s(a, (t, n) => {
				Mt(t, n > 0, e, i());
			});
		}
	}, Mt = function(e, t, n, r) {
		let i = Lt(e), a;
		for (; a = i.nextNode();) Es(r, a) && !Es(qe, a) && (t && Ut(T.uponSanitizeShadowNode, a, null), Yt(a, n), nn(a, n), L(a.content) && R(a.content));
	}, Nt = function(e) {
		qe && Ts(qe, e);
	}, Pt = function(e, t) {
		return Pe ? e === "patchsrc" || e === "for" && t !== "label" && t !== "output" : !1;
	}, Ft = function(e) {
		Pe && Et(e, (e) => {
			let t = C(e);
			if (t === $.processingInstruction || t === $.comment && X(hc, e.data)) {
				try {
					p(e);
				} catch {}
				return !1;
			}
			if (t === $.element) {
				let t = e, n = N(w(e));
				try {
					t.hasAttribute && t.hasAttribute("patchsrc") && t.removeAttribute("patchsrc"), t.hasAttribute && t.hasAttribute("for") && Pt("for", n) && t.removeAttribute("for");
				} catch {}
			}
			return !0;
		});
	}, It = function(e) {
		let t = null, r = null;
		if (ze) e = "<remove></remove>" + e;
		else {
			let t = ks(e, /^[\r\n\t ]+/);
			r = t && t[0];
		}
		ft === "application/xhtml+xml" && rt === nt && (e = "<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>" + e + "</body></html>");
		let i = ne ? ce(e) : e;
		if (rt === nt) try {
			t = new l().parseFromString(i, ft);
		} catch {}
		if (!t || !t.documentElement) {
			t = fe.createDocument(rt, "template", null);
			try {
				t.documentElement.innerHTML = it ? re : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), rt === nt ? he.call(t, Fe ? "html" : "body")[0] : Fe ? t.documentElement : a;
	}, Lt = function(e) {
		let t = S ? S(e) : e.ownerDocument;
		return pe.call(t || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, Rt = function(e) {
		return e = As(e, _e, " "), e = As(e, ve, " "), e = As(e, ye, " "), e;
	}, zt = function(e) {
		Bt(e);
		let t = e.querySelectorAll?.call(e, "template");
		t && _s(t, (e) => {
			L(e.content) && zt(e.content);
		});
	}, Bt = function(e) {
		e.normalize();
		let t = S ? S(e) : e.ownerDocument, n = pe.call(t || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), r = n.nextNode();
		for (; r;) r.data = Rt(r.data), r = n.nextNode();
	}, Vt = function(e) {
		let t = x ? x(e) : null;
		return typeof t != "string" || N(t) !== "form" ? !1 : typeof e.nodeName != "string" || typeof e.textContent != "string" || typeof e.removeChild != "function" || e.attributes !== y(e) || typeof e.removeAttribute != "function" || typeof e.removeAttributeNode != "function" || typeof e.getAttributeNode != "function" || typeof e.setAttribute != "function" || typeof e.namespaceURI != "string" || typeof e.insertBefore != "function" || typeof e.hasChildNodes != "function" || e.nodeType !== b(e) || e.childNodes !== g(e);
	}, L = function(e) {
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
		e.length !== 0 && _s(e, (e) => {
			e.call(t, n, r, P);
		});
	}
	let Wt = function(e, t) {
		return !!(Pe && e.hasChildNodes() && !Ht(e.firstElementChild) && X(mc, e.textContent) && X(mc, e.innerHTML) || Pe && vc[t] && C(e) === $.element && te(e) === nt && (Ht(e.firstElementChild) || typeof e.textContent == "string" && X(mc, e.textContent)) || e.nodeType === $.processingInstruction || Pe && e.nodeType === $.comment && X(hc, e.data));
	}, Gt = function(e, t) {
		return e instanceof RegExp ? X(e, t) : e instanceof Function && !!e(t, ...[...arguments].slice(2));
	}, Kt = function(e, t, n) {
		if (!Ee[t] && Qt(t) && Gt(A.tagNameCheck, t)) return !1;
		if (j && !Ye[t] && !Ye[Ds(t)]) {
			let t = _(e), r = g(e);
			if (r && t) {
				let i = r.length;
				for (let a = i - 1; a >= 0; --a) {
					let i = e === n ? f(r[a], !0) : r[a];
					t.insertBefore(i, h(e));
				}
			}
		}
		return I(e), !0;
	}, qt = function(e, t, n, r) {
		return e.length === 0 ? t : t === n || t === r ? Hs(t) : t;
	}, Jt = function(e, t) {
		return e === t || _(e) !== null ? !1 : (Ke && Dt(e), !0);
	}, Yt = function(e, n) {
		if (Nt(e), Ut(T.beforeSanitizeElements, e, null), Jt(e, n)) return !0;
		if (Vt(e)) return I(e), !0;
		let r = N(w(e));
		if (D = qt(T.uponSanitizeElement, D, O, Le), Ut(T.uponSanitizeElement, e, {
			tagName: r,
			allowedTags: D
		}), Jt(e, n)) return !0;
		if (Wt(e, r)) return I(e), !0;
		if (Ee[r] || !(Oe.tagCheck instanceof Function && Oe.tagCheck(r)) && !D[r]) {
			let t = Kt(e, r, n);
			return t === !1 && (Ut(T.afterSanitizeElements, e, null), Jt(e, n)) ? !0 : t;
		}
		if (C(e) === $.element && !xt(e) || (r === "noscript" || r === "noembed" || r === "noframes") && X(gc, e.innerHTML)) return I(e), !0;
		if (Ne && e.nodeType === $.text) {
			let n = Rt(e.textContent);
			e.textContent !== n && (bs(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return Ut(T.afterSanitizeElements, e, null), Jt(e, n);
	}, Xt = function(e, t, r) {
		if (De[t] || Pt(t, e) || Ue && (t === "id" || t === "name") && (r in n || r in mt)) return !1;
		let i = k[t] || Oe.attributeCheck instanceof Function && Oe.attributeCheck(t, e);
		return Ae && X(be, t) || ke && X(xe, t) ? !0 : i ? $e[t] || X(E, As(r, Ce, "")) || (t === "src" || t === "xlink:href" || t === "href") && e !== "script" && js(r, "data:") === 0 && Ze[e] || je && !X(Se, As(r, Ce, "")) ? !0 : !r : Qt(e) && Gt(A.tagNameCheck, e) && Gt(A.attributeNameCheck, t, e) || t === "is" && A.allowCustomizedBuiltInElements && Gt(A.tagNameCheck, r);
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
		return !Zt[Ds(e)] && X(we, e);
	}, $t = function(e, t, n, r) {
		if (ne && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return ce(r);
			case "TrustedScriptURL": return le(r);
		}
		return r;
	}, en = function(e, t, n, r) {
		try {
			return n ? e.setAttributeNS(n, t, r) : e.setAttribute(t, r), !Vt(e) || (I(e), !1);
		} catch {
			return wt(t, e), !1;
		}
	}, tn = function(e, t) {
		return Pe && X(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, t) ? !0 : Ds(e) === "attributename" && ks(t, "href") !== null;
	}, nn = function(e, n) {
		if (Ut(T.beforeSanitizeAttributes, e, null), Jt(e, n)) return;
		let r = e.attributes;
		if (!r || Vt(e)) return;
		k = qt(T.uponSanitizeAttribute, k, Te, Re);
		let i = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: k,
			forceKeepAttr: void 0
		}, a = r.length, o = N(e.nodeName);
		for (; a--;) {
			let n = r[a], s = n.name, c = n.namespaceURI, l = n.value, u = N(s), d = l, f = s === "value" ? d : Ms(d), p = !1;
			i.attrName = u, i.attrValue = f, i.keepAttr = !0, i.forceKeepAttr = void 0, Ut(T.uponSanitizeAttribute, e, i), f = i.attrValue, We && (u === "id" || u === "name") && js(f, Ge) !== 0 && (wt(s, e, n), f = Ge + f, p = !0), tn(s, f) ? wt(s, e, n) : i.forceKeepAttr || (i.keepAttr ? !Me && X(_c, f) ? wt(s, e, n) : (Ne && (f = Rt(f)), Xt(o, u, f) ? (f = $t(o, u, c, f), f !== d && en(e, s, c, f) && p && ys(t.removed)) : wt(s, e, n)) : wt(s, e, n));
		}
		Ut(T.afterSanitizeAttributes, e, null), Jt(e, n);
	}, R = function(e) {
		let t = null, n = Lt(e);
		for (Ut(T.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (Ut(T.uponSanitizeShadowNode, t, null), Yt(t, e), nn(t, e), L(t.content) && R(t.content), C(t) === $.element) {
			let e = v(t);
			L(e) && (z(e), R(e));
		}
		Ne && Bt(e), Ut(T.afterSanitizeShadowDOM, e, null);
	}, z = function(e) {
		let t = [{
			node: e,
			shadow: null
		}];
		for (; t.length > 0;) {
			let e = t.pop();
			if (e.shadow) {
				R(e.shadow);
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
					L(e) && t.push({
						node: e,
						shadow: null
					});
				}
			}
			if (r) {
				let e = v(n);
				L(e) && t.push({
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
		if (it = !e, it && (e = "<!-->"), typeof e != "string" && !Ht(e) && (e = Us(e), typeof e != "string")) throw zs("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		Ie ? (D = Le, k = Re) : gt(n), (T.uponSanitizeElement.length > 0 || T.uponSanitizeAttribute.length > 0) && (D = Hs(D)), T.uponSanitizeAttribute.length > 0 && (k = Hs(k)), t.removed = [];
		let c = Ke && typeof e != "string" && Ht(e), l = [], u = qe;
		if (c) {
			Ft(e);
			let t = w(e);
			if (typeof t == "string") {
				let n = N(t);
				if (!D[n] || Ee[n]) throw Ct(e), zs("root node is forbidden and cannot be sanitized in-place");
			}
			if (Vt(e)) throw Ct(e), zs("root node is clobbered and cannot be sanitized in-place");
			l = Ot(e), qe = ws();
		} else if (Ht(e)) i = It("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === $.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), z(i);
		else {
			if (!Be && !Ne && !Fe && e.indexOf("<") === -1) return ne && He ? ce(e) : e;
			if (i = It(e), !i) return Be ? null : He ? re : "";
		}
		i && ze && I(i.firstChild);
		let d = c ? e : i;
		try {
			c && z(e);
			let t = Lt(d);
			for (; o = t.nextNode();) Yt(o, d), nn(o, d), L(o.content) && R(o.content);
			c && jt(e, l);
		} catch (t) {
			throw c && (Ct(e), At(e, l)), t;
		} finally {
			qe = u;
		}
		if (c) {
			if (At(e, l)) throw zs("a node selected for removal could not be safely returned; refusing to sanitize in place");
			return Ne && zt(e), e;
		}
		if (Be) {
			if (Ne && zt(i), Ve) for (s = me.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (k.shadowroot || k.shadowrootmode) && (s = ge.call(r, s, !0)), s;
		}
		let f = Fe ? i.outerHTML : i.innerHTML;
		return Fe && D["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && X(fc, i.ownerDocument.doctype.name) && (f = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + f), Ne && (f = Rt(f)), ne && He ? ce(f) : f;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		gt(e), Ie = !0, Le = D, Re = k;
	}, t.clearConfig = function() {
		P = null, Ie = !1, Le = null, Re = null, ne = ie, re = "";
	}, t.isValidAttribute = function(e, t, n) {
		P || gt({});
		let r = N(e), i = N(t);
		return typeof n == "string" && (tn(t, n) || !Me && X(_c, n)) ? !1 : Xt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && Ls(T, e) && bs(T[e], t);
	}, t.removeHook = function(e, t) {
		if (Ls(T, e)) {
			if (t !== void 0) {
				let n = vs(T[e], t);
				return n === -1 ? void 0 : Ss(T[e], n, 1)[0];
			}
			return ys(T[e]);
		}
	}, t.removeHooks = function(e) {
		Ls(T, e) && (T[e] = []);
	}, t.removeAllHooks = function() {
		T = xc();
	}, t;
}
wc();
// kit/format.js
function Tc(e) {
	if (!e || e.startsWith("0001")) return "";
	let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3), n = Math.abs(t), r = t < 0 ? "in " : "", i = t < 0 ? "" : " ago";
	return n < 60 ? r + n + "s" + i : n < 3600 ? r + Math.round(n / 60) + "m" + i : n < 86400 ? r + Math.round(n / 3600) + "h" + i : r + Math.round(n / 86400) + "d" + i;
}
function Ec(e) {
	try {
		let t = new URL(e);
		return t.protocol === "https:" || t.protocol === "http:";
	} catch {
		return !1;
	}
}
// node_modules/@lucide/svelte/dist/utils/defaultAttributes.js
var Dc = {
	xmlns: "http://www.w3.org/2000/svg",
	width: 24,
	height: 24,
	viewBox: "0 0 24 24",
	fill: "none",
	stroke: "currentColor",
	"stroke-width": 2,
	"stroke-linecap": "round",
	"stroke-linejoin": "round"
}, Oc = (...e) => e.filter((e, t, n) => !!e && e.trim() !== "" && n.indexOf(e) === t).join(" ").trim();
// node_modules/@lucide/svelte/dist/utils/buildLucideIconNode.js
function kc(e) {
	return e != null;
}
function Ac(e, t = {}) {
	let n = t.attributeNames ?? {}, r = (e) => n[e] ?? e, i = e.size ?? e.width ?? Dc.width, a = e.size ?? e.height ?? Dc.height, o = e.aliases?.filter((e) => typeof e == "string" && e.trim() !== "").map((e) => `lucide-${e}`) ?? [], s = [...e.name ? [`lucide-${e.name}`] : [], ...o], c = t.className?.split(" ").filter(Boolean) ?? [], l = t.includeDefaultClasses === !1 ? Oc(...c) : Oc("lucide", ...s, ...c), u = t.absoluteStrokeWidth ? Number(t.strokeWidth ?? Dc["stroke-width"]) * Number(e.size ?? e.width ?? Dc.width) / Number(t.size ?? t.width ?? Dc.width) : t.strokeWidth ?? Dc["stroke-width"];
	return [
		"svg",
		{
			...Object.entries(Dc).reduce((e, [t, n]) => (e[r(t)] = n, e), {}),
			..."color" in t && t.color && { [r("stroke")]: t.color },
			..."size" in t && kc(t.size) && {
				[r("width")]: t.size,
				[r("height")]: t.size
			},
			..."width" in t && kc(t.width) && { [r("width")]: t.width },
			..."height" in t && kc(t.height) && { [r("height")]: t.height },
			[r("stroke-width")]: u,
			...l && { [r("class")]: l },
			[r("viewBox")]: `0 0 ${i} ${a}`,
			...t.hasA11yProp === !1 ? { [r("aria-hidden")]: "true" } : {},
			..."attributes" in t && t.attributes
		},
		e.node.map((e) => {
			let [n, i, a] = e, o = t.nonScalingStroke ? {
				[r("vector-effect")]: "non-scaling-stroke",
				...i
			} : i;
			return a ? [
				n,
				o,
				a
			] : [n, o];
		})
	];
}
// node_modules/@lucide/svelte/dist/utils/hasA11yProp.js
var jc = (e) => {
	for (let t in e) if (t.startsWith("aria-") || t === "role" || t === "title") return !0;
	return !1;
}, Mc = Symbol("lucide-context"), Nc = () => qe(Mc), Pc = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy",
	"color",
	"size",
	"width",
	"height",
	"strokeWidth",
	"absoluteStrokeWidth",
	"nonScalingStroke",
	"iconNode",
	"icon",
	"class",
	"children"
]), Fc = /* @__PURE__ */ Dr("<svg><!><!></svg>");
function Ic(e, t) {
	Je(t, !0);
	let n = Nc() ?? {}, r = Zi(t, "color", 19, () => n.color ?? "currentColor"), i = Zi(t, "size", 19, () => n.size ?? 24), a = Zi(t, "width", 19, i), o = Zi(t, "height", 19, i), s = Zi(t, "strokeWidth", 19, () => n.strokeWidth ?? 2), c = Zi(t, "absoluteStrokeWidth", 19, () => n.absoluteStrokeWidth ?? !1), l = Zi(t, "nonScalingStroke", 19, () => n.nonScalingStroke ?? !1), u = Zi(t, "iconNode", 19, () => []), d = Zi(t, "icon", 19, () => ({
		node: u(),
		aliases: [],
		size: 24
	})), f = /* @__PURE__ */ Ji(t, Pc), p = /* @__PURE__ */ P(() => !!t.children || jc(f)), m = /* @__PURE__ */ P(() => Ac(d(), {
		color: r(),
		width: a(),
		height: o(),
		strokeWidth: s(),
		absoluteStrokeWidth: c(),
		nonScalingStroke: l(),
		className: Oc("lucide-icon", n.class),
		hasA11yProp: W(p),
		attributes: f
	})), h = /* @__PURE__ */ P(() => b(W(m), 3)), _ = /* @__PURE__ */ P(() => W(h)[1]), v = /* @__PURE__ */ P(() => y(W(h)[2], () => [], !0)), x = /* @__PURE__ */ P(() => ({
		...W(_),
		class: [...W(_).class.split(" "), t.class]
	}));
	var S = Fc();
	Vi(S, () => ({ ...W(x) }));
	var C = an(S);
	ni(C, 17, () => W(v), Qr, (e, t) => {
		var n = /* @__PURE__ */ P(() => b(W(t), 2));
		let r = () => W(n)[0], i = () => W(n)[1];
		var a = Or();
		li(on(a), r, !0, (e, t) => {
			Vi(e, () => ({ ...i() }));
		}), G(e, a);
	}), Xr(B(C), () => t.children ?? g), A(S), G(e, S), Ye();
}
// node_modules/@lucide/svelte/dist/icons/git-commit-horizontal.svelte
var Lc = {
	name: "git-commit-horizontal",
	size: 24,
	node: [
		["circle", {
			cx: "12",
			cy: "12",
			r: "3"
		}],
		["line", {
			x1: "3",
			x2: "9",
			y1: "12",
			y2: "12"
		}],
		["line", {
			x1: "15",
			x2: "21",
			y1: "12",
			y2: "12"
		}]
	],
	aliases: ["git-commit"]
};
Lc.node;
var Rc = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function zc(e, t) {
	let n = /* @__PURE__ */ Ji(t, Rc);
	Ic(e, Xi(() => n, { get icon() {
		return Lc;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/git-pull-request.svelte
var Bc = {
	name: "git-pull-request",
	size: 24,
	node: [
		["circle", {
			cx: "18",
			cy: "18",
			r: "3"
		}],
		["circle", {
			cx: "6",
			cy: "6",
			r: "3"
		}],
		["path", { d: "M13 6h3a2 2 0 0 1 2 2v7" }],
		["line", {
			x1: "6",
			x2: "6",
			y1: "9",
			y2: "21"
		}]
	]
};
Bc.node;
var Vc = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Hc(e, t) {
	let n = /* @__PURE__ */ Ji(t, Vc);
	Ic(e, Xi(() => n, { get icon() {
		return Bc;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/git-merge.svelte
var Uc = {
	name: "git-merge",
	size: 24,
	node: [
		["circle", {
			cx: "18",
			cy: "18",
			r: "3"
		}],
		["circle", {
			cx: "6",
			cy: "6",
			r: "3"
		}],
		["path", { d: "M6 21V9a9 9 0 0 0 9 9" }]
	]
};
Uc.node;
var Wc = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Gc(e, t) {
	let n = /* @__PURE__ */ Ji(t, Wc);
	Ic(e, Xi(() => n, { get icon() {
		return Uc;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/git-branch.svelte
var Kc = {
	name: "git-branch",
	size: 24,
	node: [
		["path", { d: "M15 6a9 9 0 0 0-9 9V3" }],
		["circle", {
			cx: "18",
			cy: "6",
			r: "3"
		}],
		["circle", {
			cx: "6",
			cy: "18",
			r: "3"
		}]
	]
};
Kc.node;
var qc = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Jc(e, t) {
	let n = /* @__PURE__ */ Ji(t, qc);
	Ic(e, Xi(() => n, { get icon() {
		return Kc;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/message-square.svelte
var Yc = {
	name: "message-square",
	size: 24,
	node: [["path", { d: "M22 17a2 2 0 0 1-2 2H6.828a2 2 0 0 0-1.414.586l-2.202 2.202A.71.71 0 0 1 2 21.286V5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2z" }]]
};
Yc.node;
var Xc = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Zc(e, t) {
	let n = /* @__PURE__ */ Ji(t, Xc);
	Ic(e, Xi(() => n, { get icon() {
		return Yc;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/message-square-code.svelte
var Qc = {
	name: "message-square-code",
	size: 24,
	node: [
		["path", { d: "M22 17a2 2 0 0 1-2 2H6.828a2 2 0 0 0-1.414.586l-2.202 2.202A.71.71 0 0 1 2 21.286V5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2z" }],
		["path", { d: "m10 8-3 3 3 3" }],
		["path", { d: "m14 14 3-3-3-3" }]
	]
};
Qc.node;
var $c = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function el(e, t) {
	let n = /* @__PURE__ */ Ji(t, $c);
	Ic(e, Xi(() => n, { get icon() {
		return Qc;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/eye.svelte
var tl = {
	name: "eye",
	size: 24,
	node: [["path", { d: "M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0" }], ["circle", {
		cx: "12",
		cy: "12",
		r: "3"
	}]]
};
tl.node;
var nl = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function rl(e, t) {
	let n = /* @__PURE__ */ Ji(t, nl);
	Ic(e, Xi(() => n, { get icon() {
		return tl;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/circle-check.svelte
var il = {
	name: "circle-check",
	size: 24,
	node: [["circle", {
		cx: "12",
		cy: "12",
		r: "10"
	}], ["path", { d: "m16 9-5.5 5.5L8 12" }]],
	aliases: ["check-circle-2"]
};
il.node;
var al = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function ol(e, t) {
	let n = /* @__PURE__ */ Ji(t, al);
	Ic(e, Xi(() => n, { get icon() {
		return il;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/circle-x.svelte
var sl = {
	name: "circle-x",
	size: 24,
	node: [
		["circle", {
			cx: "12",
			cy: "12",
			r: "10"
		}],
		["path", { d: "m15 9-6 6" }],
		["path", { d: "m9 9 6 6" }]
	],
	aliases: ["x-circle"]
};
sl.node;
var cl = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function ll(e, t) {
	let n = /* @__PURE__ */ Ji(t, cl);
	Ic(e, Xi(() => n, { get icon() {
		return sl;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/circle-dot.svelte
var ul = {
	name: "circle-dot",
	size: 24,
	node: [["circle", {
		cx: "12",
		cy: "12",
		r: "1"
	}], ["circle", {
		cx: "12",
		cy: "12",
		r: "10"
	}]]
};
ul.node;
var dl = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function fl(e, t) {
	let n = /* @__PURE__ */ Ji(t, dl);
	Ic(e, Xi(() => n, { get icon() {
		return ul;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/tag.svelte
var pl = {
	name: "tag",
	size: 24,
	node: [["path", { d: "M12.586 2.586A2 2 0 0 0 11.172 2H4a2 2 0 0 0-2 2v7.172a2 2 0 0 0 .586 1.414l8.704 8.704a2.426 2.426 0 0 0 3.42 0l6.58-6.58a2.426 2.426 0 0 0 0-3.42z" }], ["circle", {
		cx: "7.5",
		cy: "7.5",
		r: ".5",
		fill: "currentColor"
	}]]
};
pl.node;
var ml = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function hl(e, t) {
	let n = /* @__PURE__ */ Ji(t, ml);
	Ic(e, Xi(() => n, { get icon() {
		return pl;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/refresh-cw.svelte
var gl = {
	name: "refresh-cw",
	size: 24,
	node: [
		["path", { d: "M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8" }],
		["path", { d: "M21 3v5h-5" }],
		["path", { d: "M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16" }],
		["path", { d: "M8 16H3v5" }]
	]
};
gl.node;
var _l = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function vl(e, t) {
	let n = /* @__PURE__ */ Ji(t, _l);
	Ic(e, Xi(() => n, { get icon() {
		return gl;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/user-round.svelte
var yl = {
	name: "user-round",
	size: 24,
	node: [["circle", {
		cx: "12",
		cy: "8",
		r: "5"
	}], ["path", { d: "M20 21a8 8 0 0 0-16 0" }]],
	aliases: ["user-2"]
};
yl.node;
var bl = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function xl(e, t) {
	let n = /* @__PURE__ */ Ji(t, bl);
	Ic(e, Xi(() => n, { get icon() {
		return yl;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/activity.svelte
var Sl = {
	name: "activity",
	size: 24,
	node: [["path", { d: "M22 12h-2.48a2 2 0 0 0-1.93 1.46l-2.35 8.36a.25.25 0 0 1-.48 0L9.24 2.18a.25.25 0 0 0-.48 0l-2.35 8.36A2 2 0 0 1 4.49 12H2" }]]
};
Sl.node;
var Cl = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function wl(e, t) {
	let n = /* @__PURE__ */ Ji(t, Cl);
	Ic(e, Xi(() => n, { get icon() {
		return Sl;
	} }));
}
// plugins/activity/Activity.svelte
var Tl = /* @__PURE__ */ Tr("<button> </button>"), El = /* @__PURE__ */ Tr("<option> </option>"), Dl = /* @__PURE__ */ Tr("<div class=\"card err\"> </div>"), Ol = /* @__PURE__ */ Tr("<div class=\"card empty faint svelte-dcm8ng\">Activity is off. Set <span class=\"mono\"></span> in jobshop.yaml.</div>"), kl = /* @__PURE__ */ Tr("<span class=\"pill\">bot</span>"), Al = /* @__PURE__ */ Tr("<div class=\"body muted svelte-dcm8ng\"> </div>"), jl = /* @__PURE__ */ Tr("<a target=\"_blank\" rel=\"noopener noreferrer\" class=\"svelte-dcm8ng\">open</a>"), Ml = /* @__PURE__ */ Tr("<a class=\"mono svelte-dcm8ng\"> </a>"), Nl = /* @__PURE__ */ Tr("<li class=\"svelte-dcm8ng\"><span><!></span> <div class=\"main svelte-dcm8ng\"><div class=\"line svelte-dcm8ng\"><span class=\"actor svelte-dcm8ng\"> </span> <!> <span class=\"title svelte-dcm8ng\"> </span></div> <!> <div class=\"meta faint svelte-dcm8ng\"><span class=\"mono\"> </span> <span> </span> <!> <!></div></div></li>"), Pl = /* @__PURE__ */ Tr("<h2 class=\"st\"> </h2> <ul class=\"card feed svelte-dcm8ng\" data-testid=\"activity\"></ul>", 1), Fl = /* @__PURE__ */ Tr("<div class=\"card empty faint svelte-dcm8ng\">Nothing in the last 7 days for these filters.</div>"), Il = /* @__PURE__ */ Tr("<!> <!>", 1), Ll = /* @__PURE__ */ Tr("<header class=\"vh svelte-dcm8ng\"><h1>Activity</h1> <div class=\"chips actions svelte-dcm8ng\" data-testid=\"activity-filters\"><!> <span class=\"sep\"></span> <button>all sources</button> <!> <span class=\"sep\"></span> <select aria-label=\"kind\"></select></div></header> <div class=\"pad\"><!></div>", 1);
function Rl(e, t) {
	Je(t, !0);
	let n = ea(() => t.api), r = (e) => t.api.go(e), i = [
		"",
		"push",
		"pr",
		"merge",
		"review",
		"review_comment",
		"comment",
		"issue",
		"ci",
		"branch",
		"release",
		"state",
		"assignee"
	], a = [
		"humans",
		"bots",
		"all"
	], o = /* @__PURE__ */ P(() => n.who || "humans"), s = $i(() => t.api, () => {
		let e = new URLSearchParams({
			since: "168h",
			limit: "300"
		});
		return n.source && e.set("source", n.source), n.repo && e.set("repo", n.repo), n.kind && e.set("kind", n.kind), W(o) !== "all" && e.set("bots", W(o)), "api/activity?" + e.toString();
	}), c = (e) => {
		let t = e.indexOf(":");
		return {
			source: e.slice(0, t),
			repo: e.slice(t + 1)
		};
	};
	function l(e) {
		switch (e.kind) {
			case "push": return zc;
			case "pr": return Hc;
			case "merge": return Gc;
			case "branch": return Jc;
			case "comment": return Zc;
			case "review_comment": return el;
			case "review": return rl;
			case "ci": return / success /.test(e.title) ? ol : ll;
			case "issue": return fl;
			case "release": return hl;
			case "state": return vl;
			case "assignee": return xl;
			default: return wl;
		}
	}
	let u = (e) => e.kind === "ci" ? / success /.test(e.title) ? "ok" : / (failure|timed_out|startup_failure) /.test(e.title) ? "bad" : "" : "";
	function d(e) {
		if (!e) return "";
		let t = e.replace(/\r/g, "").trim();
		return t.length > 280 ? t.slice(0, 280) + "…" : t;
	}
	function f(e) {
		let t = new Date(e), n = /* @__PURE__ */ new Date(), r = new Date(n);
		return r.setDate(n.getDate() - 1), t.toDateString() === n.toDateString() ? "Today" : t.toDateString() === r.toDateString() ? "Yesterday" : t.toLocaleDateString(void 0, {
			weekday: "short",
			month: "short",
			day: "numeric"
		});
	}
	let p = /* @__PURE__ */ P(() => {
		let e = [];
		for (let t of s.data?.items ?? []) {
			let n = f(t.at);
			e.length && e[e.length - 1].day === n || e.push({
				day: n,
				items: []
			}), e[e.length - 1].items.push(t);
		}
		return e;
	});
	var m = Ll(), h = on(m), g = B(an(h), 2), _ = an(g);
	ni(_, 16, () => a, (e) => e, (e, t) => {
		var n = Tl();
		let i;
		var a = sn(n, !0);
		Tn(() => {
			i = yi(n, 1, "", null, i, { on: W(o) === t }), Ur(a, t);
		}), gr("click", n, () => r({ who: t === "humans" ? "" : t })), G(e, n);
	});
	var v = B(_, 4);
	let y;
	var b = B(v, 2);
	ni(b, 16, () => s.data?.sources ?? [], (e) => e, (e, t) => {
		let i = /* @__PURE__ */ P(() => c(t));
		var a = Tl();
		let o;
		var s = sn(a, !0);
		Tn(() => {
			o = yi(a, 1, "", null, o, { on: n.repo === W(i).repo }), Ur(s, W(i).repo);
		}), gr("click", a, () => r({
			source: W(i).source,
			repo: W(i).repo
		})), G(e, a);
	});
	var x = B(b, 4);
	ni(x, 20, () => i, (e) => e, (e, t) => {
		var n = El(), r = sn(n, !0), i = {};
		Tn(() => {
			Ur(r, t || "all kinds"), i !== (i = t) && (n.value = (n.__value = i) ?? "");
		}), G(e, n);
	}), A(x);
	var S;
	Ei(x), A(g), A(h);
	var C = B(h, 2), w = an(C), ee = (e) => {
		var t = Dl(), n = sn(t, !0);
		Tn(() => Ur(n, s.error)), G(e, t);
	}, te = (e) => {
		var t = Il(), n = on(t), i = (e) => {
			var t = Ol(), n = B(an(t));
			n.textContent = "activity: {}", Ee(), A(t), G(e, t);
		};
		Zr(n, (e) => {
			s.data.sources.length || e(i);
		}), ni(B(n, 2), 17, () => W(p), (e) => e.day, (e, t) => {
			var n = Pl(), i = on(n), a = sn(i, !0), o = B(i, 2);
			ni(o, 21, () => W(t).items, (e) => e.id, (e, t) => {
				let n = /* @__PURE__ */ P(() => l(W(t)));
				var i = Nl(), a = an(i);
				ci(an(a), () => W(n), (e, t) => {
					t(e, { size: 14 });
				}), A(a);
				var o = B(a, 2), s = an(o), c = an(s), f = sn(c, !0), p = B(c, 2), m = (e) => {
					G(e, kl());
				};
				Zr(p, (e) => {
					W(t).bot && e(m);
				});
				var h = sn(B(p, 2), !0);
				A(s);
				var g = B(s, 2), _ = (e) => {
					var n = Al(), r = sn(n, !0);
					Tn((e) => Ur(r, e), [() => d(W(t).body)]), G(e, n);
				};
				Zr(g, (e) => {
					W(t).body && e(_);
				});
				var v = B(g, 2), y = an(v), b = sn(y, !0), x = B(y, 2), S = sn(x, !0), C = B(x, 2), w = (e) => {
					var n = jl();
					Tn(() => zi(n, "href", W(t).url)), G(e, n);
				}, ee = /* @__PURE__ */ P(() => W(t).url && Ec(W(t).url));
				Zr(C, (e) => {
					W(ee) && e(w);
				}), ni(B(C, 2), 16, () => W(t).tickets ?? [], (e) => e, (e, t) => {
					var n = Ml(), i = sn(n, !0);
					Tn(() => {
						zi(n, "href", "?ticket=" + t), Ur(i, t);
					}), gr("click", n, (e) => {
						e.preventDefault(), r({ ticket: t });
					}), G(e, n);
				}), A(v), A(o), A(i), Tn((e, n) => {
					yi(a, 1, `icon ${e ?? ""}`, "svelte-dcm8ng"), zi(a, "title", W(t).kind), Ur(f, W(t).actor || "unknown"), Ur(h, W(t).title), Ur(b, W(t).repo), zi(x, "title", W(t).at), Ur(S, n);
				}, [() => u(W(t)), () => Tc(W(t).at)]), G(e, i);
			}), A(o), Tn(() => Ur(a, W(t).day)), G(e, n);
		}, (e) => {
			var t = Or(), n = on(t), r = (e) => {
				G(e, Fl());
			};
			Zr(n, (e) => {
				s.data.sources.length && e(r);
			}), G(e, t);
		}), G(e, t);
	};
	Zr(w, (e) => {
		s.error ? e(ee) : s.data && e(te, 1);
	}), A(C), Tn(() => {
		y = yi(v, 1, "", null, y, { on: !n.source && !n.repo }), S !== (S = n.kind) && (x.value = (x.__value = S) ?? "", Ti(x, S));
	}), gr("click", v, () => r({
		source: "",
		repo: ""
	})), gr("change", x, (e) => r({ kind: e.currentTarget.value })), G(e, m), Ye();
}
// plugins/activity/index.js
_r(["click", "change"]), Qi("desk-activity-view", Rl);
