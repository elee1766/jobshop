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
var x = 1 << 24, S = 1024, C = 2048, w = 4096, ee = 8192, te = 16384, ne = 32768, re = 1 << 25, ie = 65536, ae = 1 << 19, oe = 1 << 20, se = 1 << 25, ce = 1 << 21, le = 1 << 22, ue = 1 << 23, T = Symbol("$state"), de = Symbol("component"), fe = Symbol("legacy props"), pe = Symbol(""), me = Symbol("attributes"), he = Symbol("class"), ge = Symbol("style"), _e = Symbol("text"), ve = Symbol("form reset"), ye = new class extends Error {
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
function Te(e) {
	E = e;
}
var D;
function O(t) {
	if (t === null) throw Se(), e;
	return D = t;
}
function Ee() {
	return O(/* @__PURE__ */ ln(D));
}
function k(t) {
	if (E) {
		if (/* @__PURE__ */ ln(D) !== null) throw Se(), e;
		D = t;
	}
}
function De(e = 1) {
	if (E) {
		for (var t = e, n = D; t--;) n = /* @__PURE__ */ ln(n);
		D = n;
	}
}
function Oe(e = !0) {
	for (var t = 0, n = D;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ ln(n);
		e && n.remove(), n = i;
	}
}
function ke(t) {
	if (!t || t.nodeType !== 8) throw Se(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function Ae(e) {
	return e === this.v;
}
function je(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function Me(e) {
	return !je(e, this.v);
}
function Ne(e) {
	throw Error("https://svelte.dev/e/lifecycle_outside_component");
}
// node_modules/svelte/src/internal/client/errors.js
function Pe() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Fe(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function Ie(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function Le() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function Re(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function ze() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Be(e) {
	throw Error("https://svelte.dev/e/props_invalid_value");
}
function Ve() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function He() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Ue() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function We() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/shared/clone.js
var Ge = [];
function Ke(e, t = !1, n = !1) {
	return qe(e, /* @__PURE__ */ new Map(), "", Ge, null, n);
}
function qe(e, t, n, r, a = null, o = !1) {
	if (typeof e == "object" && e) {
		var s = t.get(e);
		if (s !== void 0) return s;
		if (e instanceof Map) return new Map(e);
		if (e instanceof Set) return new Set(e);
		if (i(e)) {
			var c = Array(e.length);
			t.set(e, c), a !== null && t.set(a, c);
			for (var l = 0; l < e.length; l += 1) {
				var u = e[l];
				l in e && (c[l] = qe(u, t, n, r, null, o));
			}
			return c;
		}
		if (p(e) === d) {
			c = {}, t.set(e, c), a !== null && t.set(a, c);
			for (var f of Object.keys(e)) c[f] = qe(e[f], t, n, r, null, o);
			return c;
		}
		if (e instanceof Date) return e.getTime(), structuredClone(e);
		if (typeof e.toJSON == "function" && !o) return qe(e.toJSON(), t, n, r, e);
	}
	if (e instanceof EventTarget) return e;
	try {
		return structuredClone(e);
	} catch {
		return e;
	}
}
// node_modules/svelte/src/internal/shared/context.js
function Je(e) {
	let t = e.p;
	for (; t !== null && t.c === null;) t = t.p;
	return t?.c ?? null;
}
function Ye(e, t) {
	return e === null && Ne(t), e.c ??= new Map(Je(e) || void 0);
}
// node_modules/svelte/src/internal/client/context.js
var A = null;
function Xe(e) {
	A = e;
}
function Ze(e) {
	return Ye(A, "getContext").get(e);
}
function Qe(e, t = !1, n) {
	A = {
		p: A,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: W,
		l: null
	};
}
function $e(e) {
	var t = A, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) wn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, A = t.p, et(e);
}
function et(e = {}) {
	return c(e, de, { value: !0 }), e;
}
function tt() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var nt = [];
function rt() {
	var e = nt;
	nt = [], _(e);
}
function it(e) {
	if (nt.length === 0 && !kt) {
		var t = nt;
		queueMicrotask(() => {
			t === nt && rt();
		});
	}
	nt.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var at = ~(C | w | S);
function j(e, t) {
	e.f = e.f & at | t;
}
function ot(e) {
	e.f & 512 || e.deps === null ? j(e, S) : j(e, w);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function st(e, t) {
	e.f & 2048 ? t.set(e, C) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, w), j(e, S);
}
// node_modules/svelte/src/internal/client/dom/elements/misc.js
function ct(e, t) {
	if (t) {
		let t = document.body;
		e.autofocus = !0, it(() => {
			document.activeElement === t && e.focus();
		});
	}
}
var lt = !1;
function ut() {
	lt || (lt = !0, document.addEventListener("reset", (e) => {
		Promise.resolve().then(() => {
			if (!e.defaultPrevented) for (let t of e.target.elements) t[ve]?.();
		});
	}, { capture: !0 }));
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function dt(e) {
	var t = U, n = W;
	Gn(null), Kn(null);
	try {
		return e();
	} finally {
		Gn(t), Kn(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function ft(e, t, n, r) {
	let i = tt() ? gt : yt;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = W, c = pt(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				_n(e, s);
			}
			mt();
		}
	}
	var d = ht();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ vt(e))).then(u).catch((e) => _n(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), mt());
	}) : f();
}
function pt() {
	var e = W, t = U, n = A, r = N;
	return function(i = !0) {
		Kn(e), Gn(t), Xe(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function mt(e = !0) {
	Kn(null), Gn(null), Xe(null), e && N?.deactivate();
}
function ht() {
	var e = W, t = e.b, n = N, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function gt(e) {
	var n = 2 | C;
	return W !== null && (W.f |= ae), {
		ctx: A,
		deps: null,
		effects: null,
		equals: Ae,
		f: n,
		fn: e,
		reactions: null,
		rv: 0,
		v: t,
		wv: 0,
		parent: W,
		ac: null
	};
}
var _t = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function vt(e, n, r) {
	let i = W;
	i === null && Pe();
	var a = void 0, o = Gt(t), s = !U, c = /* @__PURE__ */ new Set();
	return Dn(() => {
		var t = W, n = v();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== ye && n.reject(e);
			}).finally(mt);
		} catch (e) {
			n.reject(e), mt();
		}
		var r = N;
		if (s) {
			if (t.f & 32768) var l = ht();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(_t);
			else for (let e of c.values()) e.reject(_t);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== _t && (r.activate(), t ? (o.f |= ue, Yt(o, t)) : (o.f & 8388608 && (o.f ^= ue), Yt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), Sn(() => {
		for (let e of c) e.reject(_t);
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
function M(e) {
	let t = /* @__PURE__ */ gt(e);
	return Jn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function yt(e) {
	let t = /* @__PURE__ */ gt(e);
	return t.equals = Me, t;
}
function bt(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) H(t[n]);
	}
}
function xt(e) {
	var n, r = W, i = e.parent;
	if (!Hn && i !== null && e.v !== t && i.f & 24576) return xe(), e.v;
	Kn(i);
	try {
		bt(e), n = ir(e);
	} finally {
		Kn(r);
	}
	return n;
}
function St(e) {
	var t = xt(e);
	!e.equals(t) && (e.wv = tr(), (!N?.is_fork || e.deps === null) && (N === null ? e.v = t : (N.capture(e, t, !0), Et?.capture(e, t, !0)), e.deps === null)) ? j(e, S) : Hn || (Dt === null ? ot(e) : (xn() || N?.is_fork) && Dt.set(e, t));
}
function Ct(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && dt(() => {
		t.ac.abort(ye), t.ac = null;
	}), t.fn !== null && (t.teardown = g), sr(t, 0), Mn(t));
}
function wt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && cr(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var Tt = null, N = null, Et = null, Dt = null, Ot = null, kt = !1, At = !1, jt = null, Mt = null, Nt = 0, Pt = 1, Ft = class e {
	id = Pt++;
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
		Tt === null ? Tt = this : (Tt.next = this, this.prev = Tt), Tt = this;
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
			for (var r of n.d) j(r, C), t(r);
			for (r of n.m) j(r, w), t(r);
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
		for (let [e, t] of this.#s) j(e, t), this.schedule(e);
		this.apply();
		for (var t = jt = [], n = [], r = Mt = []; this.#a.length > 0;) {
			Nt++ > 1e3 && (this.#b(), It());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Vt(e), this.#d() || this.discard(), t;
			}
		}
		if (N = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (jt = null, Mt = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Bt(e, t);
			r.length > 0 && N.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), Et = this, Rt(n), Rt(t), Et = null, this.#i?.resolve();
			var o = N;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Ut.clear(), o.#m());
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), j(i, C), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), N = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) st(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), Dt?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		N = this;
	}
	deactivate() {
		N = null, Dt = null;
	}
	flush() {
		try {
			At = !0, N = this, this.#m();
		} finally {
			Nt = 0, Ot = null, jt = null, Mt = null, At = !1, N = null, Dt = null, Ut.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(_t);
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
		this.#u || (this.#u = !0, it(() => {
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
		if (N === null) {
			let t = N = new e();
			!At && it(() => {
				t.started || t.flush();
			});
		}
		return N;
	}
	apply() {
		Dt = null;
	}
	schedule(e) {
		Ot = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? Tt = e : t.prev = e, this.linked = !1;
		}
	}
};
function It() {
	try {
		ze();
	} catch (e) {
		_n(e, Ot);
	}
}
var Lt = null;
function Rt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && nr(r) && (Lt = /* @__PURE__ */ new Set(), cr(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && Fn(r), Lt?.size > 0)) {
				Ut.clear();
				for (let e of Lt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) Lt.has(n) && (Lt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || cr(n);
					}
				}
				Lt.clear();
			}
		}
		Lt = null;
	}
}
function zt(e) {
	N.schedule(e);
}
function Bt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), j(e, S);
		for (var n = e.first; n !== null;) Bt(n, t), n = n.next;
	}
}
function Vt(e) {
	j(e, S);
	for (var t = e.first; t !== null;) Vt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Ht = /* @__PURE__ */ new Set(), Ut = /* @__PURE__ */ new Map(), Wt = !1;
function Gt(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: Ae,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function P(e, t) {
	let n = Gt(e, t);
	return Jn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Kt(e, t = !1, n = !0) {
	let r = Gt(e);
	return t || (r.equals = Me), r;
}
function F(e, t, n = !1) {
	return U !== null && (!Wn || U.f & 131072) && tt() && U.f & 4325394 && (qn === null || !qn.has(e)) && Ue(), Yt(e, n ? en(t) : t, Mt);
}
var qt = null, Jt = 0;
function Yt(e, t, n = null) {
	if (!e.equals(t)) {
		Hn ? Ut.set(e, t) : Ut.has(e) || Ut.set(e, e.v);
		var r = Ft.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && xt(t), Dt === null && ot(t);
		}
		e.wv = tr(), qt = null, Jt = 0, $t(e, C, n), qt = null, tt() && W !== null && W.f & 1024 && !(W.f & 96) && (Yn === null ? Xn([e]) : Yn.push(e)), !r.is_fork && Ht.size > 0 && !Wt && Xt();
	}
	return t;
}
function Xt() {
	Wt = !1;
	for (let e of Ht) {
		e.f & 1024 && j(e, w);
		let t;
		try {
			t = nr(e);
		} catch {
			t = !0;
		}
		t && cr(e);
	}
	Ht.clear();
}
function Zt(e, t = 1) {
	var n = q(e), r = t === 1 ? n++ : n--;
	return F(e, n), r;
}
function Qt(e) {
	F(e, e.v + 1);
}
function $t(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = tt(), a = r.length;
		if (Jt += a, Jt > 1e5 && qt === null && (qt = /* @__PURE__ */ new Set()), qt !== null) {
			if (qt.has(e)) return;
			qt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== W) {
				var l = (c & C) === 0;
				if (l && j(s, t), c & 131072) Ht.add(s);
				else if (c & 2) {
					var u = s;
					Dt?.delete(u), $t(u, w, n);
				} else if (l) {
					var d = s;
					c & 16 && Lt !== null && Lt.add(d), n === null ? zt(d) : n.push(d);
				}
			}
		}
	}
}
function en(e) {
	if (typeof e != "object" || !e || T in e || de in e) return e;
	let n = p(e);
	if (n !== d && n !== f) return e;
	var r = /* @__PURE__ */ new Map(), a = i(e), o = /* @__PURE__ */ P(0), s = null, c = $n, u = (e) => {
		if ($n === c) return e();
		var t = U, n = $n;
		Gn(null), er(c);
		var r = e();
		return Gn(t), er(n), r;
	};
	return a && r.set("length", /* @__PURE__ */ P(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Ve();
			var i = r.get(t);
			return i === void 0 ? u(() => {
				var e = /* @__PURE__ */ P(n.value, s);
				return r.set(t, e), e;
			}) : F(i, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var i = r.get(n);
			if (i === void 0) {
				if (n in e) {
					let e = u(() => /* @__PURE__ */ P(t, s));
					r.set(n, e), Qt(o);
				}
			} else F(i, t), Qt(o);
			return !0;
		},
		get(n, i, a) {
			if (i === T) return e;
			var o = r.get(i), c = i in n;
			if (o === void 0 && (!c || l(n, i)?.writable) && (o = u(() => /* @__PURE__ */ P(en(c ? n[i] : t), s)), r.set(i, o)), o !== void 0) {
				var d = q(o);
				return d === t ? void 0 : d;
			}
			return Reflect.get(n, i, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var i = Reflect.getOwnPropertyDescriptor(e, n), a = r.get(n);
			if (a !== void 0) {
				var o = q(a);
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
			if (n === T) return !0;
			var i = r.get(n), a = i !== void 0 && i.v !== t || Reflect.has(e, n);
			return (i !== void 0 || (W !== null || U !== null) && (!a || l(e, n)?.writable)) && (i === void 0 && (i = u(() => /* @__PURE__ */ P(a ? en(e[n]) : t, s)), r.set(n, i)), q(i) === t) ? !1 : a;
		},
		set(e, n, i, c) {
			var d = r.get(n), f = n in e;
			if (a && n === "length") for (var p = i; p < d.v; p += 1) {
				var m = r.get(p + "");
				m === void 0 ? p in e && (m = u(() => /* @__PURE__ */ P(t, s)), r.set(p + "", m)) : F(m, t);
			}
			if (d === void 0) (!f || l(e, n)?.writable) && (d = u(() => /* @__PURE__ */ P(void 0, s)), F(d, en(i)), r.set(n, d));
			else {
				f = d.v !== t;
				var h = u(() => en(i));
				F(d, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, i), !f) {
				if (a && typeof n == "string") {
					var _ = r.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && F(_, v + 1);
				}
				Qt(o);
			}
			return !0;
		},
		ownKeys(e) {
			q(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = r.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [i, a] of r) a.v !== t && !(i in e) && n.push(i);
			return n;
		},
		setPrototypeOf() {
			He();
		}
	});
}
function tn(e) {
	try {
		if (typeof e == "object" && e && T in e) return e[T];
	} catch {}
	return e;
}
function nn(e, t) {
	return Object.is(tn(e), tn(t));
}
var rn, an, on, sn;
function cn() {
	if (rn === void 0) {
		rn = window, an = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		on = l(t, "firstChild").get, sn = l(t, "nextSibling").get, m(e) && (e[he] = void 0, e[me] = null, e[ge] = void 0, e.__e = void 0), m(n) && (n[_e] = void 0);
	}
}
function I(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function L(e) {
	return on.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function ln(e) {
	return sn.call(e);
}
function R(e, t) {
	if (!E) return /* @__PURE__ */ L(e);
	var n = /* @__PURE__ */ L(D);
	if (n === null) n = D.appendChild(I());
	else if (t && n.nodeType !== 3) {
		var r = I();
		return n?.before(r), O(r), r;
	}
	return t && hn(n), O(n), n;
}
function un(e, t = !1) {
	if (!E) {
		var n = /* @__PURE__ */ L(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ ln(n) : n;
	}
	if (t) {
		if (D?.nodeType !== 3) {
			var r = I();
			return D?.before(r), O(r), r;
		}
		hn(D);
	}
	return D;
}
function dn(e, t = !1) {
	if (!E) return /* @__PURE__ */ L(e);
	var n = R(e, t);
	return k(e), n;
}
function z(e, t = 1, n = !1) {
	let r = E ? D : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ ln(r);
	if (!E) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = I();
			return r === null ? i?.after(a) : r.before(a), O(a), a;
		}
		hn(r);
	}
	return O(r), r;
}
function fn(e) {
	e.textContent = "";
}
function pn() {
	return !1;
}
function mn(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function hn(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function gn(e) {
	var t = W;
	if (t === null) return U.f |= ue, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	_n(e, t);
}
function _n(t, n) {
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
function vn(e) {
	W === null && (U === null && Re(e), Le()), Hn && Ie(e);
}
function yn(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function bn(e, t) {
	var n = W;
	n !== null && n.f & 8192 && (e |= ee);
	var r = {
		ctx: A,
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
	N?.register_created_effect(r);
	var i = r;
	if (e & 4) jt === null ? Ft.ensure().schedule(r) : jt.push(r);
	else if (t !== null) {
		try {
			cr(r);
		} catch (e) {
			throw H(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= ie));
	}
	if (i !== null && (i.parent = n, n !== null && yn(i, n), U !== null && U.f & 2 && !(e & 64))) {
		var a = U;
		(a.effects ??= []).push(i);
	}
	return r;
}
function xn() {
	return U !== null && !Wn;
}
function Sn(e) {
	let t = bn(8, null);
	return j(t, S), t.teardown = e, t;
}
function Cn(e) {
	vn("$effect");
	var t = W.f;
	if (!U && t & 32 && A !== null && !A.i) {
		var n = A;
		(n.e ??= []).push(e);
	} else return wn(e);
}
function wn(e) {
	return bn(4 | oe, e);
}
function Tn(e) {
	Ft.ensure();
	let t = bn(64 | ae, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? In(t, () => {
			H(t), n(void 0);
		}) : (H(t), n(void 0));
	});
}
function En(e) {
	return bn(4, e);
}
function Dn(e) {
	return bn(le | ae, e);
}
function On(e, t = 0) {
	return bn(8 | t, e);
}
function B(e, t = [], n = [], r = []) {
	ft(r, t, n, (t) => {
		bn(8, () => {
			e(...t.map(q));
		});
	});
}
function kn(e, t = 0) {
	return bn(16 | t, e);
}
function An(e, t = 0) {
	return bn(x | t, e);
}
function V(e) {
	return bn(32 | ae, e);
}
function jn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = Hn, r = U;
		Un(!0), Gn(null);
		try {
			t.call(null);
		} catch (t) {
			_n(t, e.parent);
		} finally {
			Un(n), Gn(r);
		}
	}
}
function Mn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && dt(() => {
			e.abort(ye);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : H(n, t), n = r;
	}
}
function Nn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || H(t), t = n;
	}
}
function H(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (Pn(e.nodes.start, e.nodes.end), n = !0), e.f |= re, Mn(e, t && !n), sr(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	jn(e), e.f ^= re, e.f |= te;
	var i = e.parent;
	i !== null && i.first !== null && Fn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function Pn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ ln(e);
		e.remove(), e = n;
	}
}
function Fn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function In(e, t, n = !0) {
	var r = [];
	e.f |= 256, Ln(e, r, !0);
	var i = () => {
		n && H(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function Ln(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= ee;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				Ln(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function Rn(e) {
	e.f &= -257, zn(e, !0);
}
function zn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= ee, e.f & 1024 || (j(e, C), Ft.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			zn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function Bn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ ln(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var Vn = null, Hn = !1;
function Un(e) {
	Hn = e;
}
var U = null, Wn = !1;
function Gn(e) {
	U = e;
}
var W = null;
function Kn(e) {
	W = e;
}
var qn = null;
function Jn(e) {
	U !== null && (U.f & 2097152 || U.f & 2) && (qn ??= /* @__PURE__ */ new Set()).add(e);
}
var G = null, K = 0, Yn = null;
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
			if (nr(a) && St(a), a.wv > e.wv) return !0;
		}
		t & 512 && Dt === null && j(e, S);
	}
	return !1;
}
function rr(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(qn !== null && qn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? rr(a, t, !1) : t === a && (n ? j(a, C) : a.f & 1024 && j(a, w), zt(a));
	}
}
function ir(e) {
	var t = G, n = K, r = Yn, i = U, a = qn, o = A, s = Wn, c = $n, l = e.f;
	G = null, K = 0, Yn = null, U = l & 96 ? null : e, qn = null, Xe(e.ctx), Wn = !1, $n = ++Qn, e.ac !== null && (dt(() => {
		e.ac.abort(ye);
	}), e.ac = null);
	try {
		e.f |= ce;
		var u = e.fn, d = u();
		e.f |= ne;
		var f = ar(e);
		if (tt() && Yn !== null && !Wn && f !== null && !(e.f & 6146)) for (var p = 0; p < Yn.length; p++) rr(Yn[p], e);
		if (i !== null && i !== e) {
			if (Qn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Qn;
			if (t !== null) for (let e of t) e.rv = Qn;
			Yn !== null && (r === null ? r = Yn : r.push(...Yn));
		}
		return e.f & 8388608 && (e.f ^= ue), d;
	} catch (t) {
		return ar(e), gn(t);
	} finally {
		e.f ^= ce, G = t, K = n, Yn = r, U = i, qn = a, Xe(o), Wn = s, $n = c;
	}
}
function ar(e) {
	var t = e.deps, n = N?.is_fork;
	if (G !== null) {
		var r;
		if (n || sr(e, K), t !== null && K > 0) for (t.length = K + G.length, r = 0; r < G.length; r++) t[K + r] = G[r];
		else e.deps = t = G;
		if (xn() && e.f & 512) for (r = K; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && K < t.length && (sr(e, K), t.length = K);
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
	if (r === null && n.f & 2 && (G === null || !o.call(G, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && ot(c), c.ac !== null && dt(() => {
			c.ac.abort(ye), c.ac = null, j(c, C);
		}), Ct(c), sr(c, 0);
	}
}
function sr(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) or(e, n[r]);
}
function cr(e) {
	var t = e.f;
	if (!(t & 16384)) {
		j(e, S);
		var n = W;
		W = e;
		try {
			t & 16777232 ? Nn(e) : Mn(e), jn(e);
			var r = ir(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Zn;
		} finally {
			W = n;
		}
	}
}
function q(e) {
	var t = !!(e.f & 2);
	if (Vn?.add(e), U !== null && !Wn && !(W !== null && W.f & 16384) && (qn === null || !qn.has(e))) {
		var n = U.deps;
		if (U.f & 2097152) e.rv < Qn && (e.rv = Qn, G === null && n !== null && n[K] === e ? K++ : G === null ? G = [e] : G.push(e));
		else {
			U.deps ??= [], o.call(U.deps, e) || U.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [U] : o.call(r, U) || r.push(U);
		}
	}
	if (Hn && Ut.has(e)) return Ut.get(e);
	if (t) {
		var i = e;
		if (Hn) {
			var a = i.v;
			return (!(i.f & 1024) && i.reactions !== null || ur(i)) && (a = xt(i)), Ut.set(i, a), a;
		}
		var s = !(i.f & 512) && !Wn && U !== null && !!(U.f & 512), c = (i.f & ne) === 0;
		nr(i) && (s && (i.f |= 512), St(i)), s && !c && (wt(i), lr(i));
	}
	if (Dt?.has(e)) return Dt.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function lr(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : o.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (wt(n), lr(n));
	}
}
function ur(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Ut.has(t) || t.f & 2 && ur(t)) return !0;
	return !1;
}
function dr(e) {
	var t = Wn;
	try {
		return Wn = !0, e();
	} finally {
		Wn = t;
	}
}
function fr(e) {
	if (!(typeof e != "object" || !e || e instanceof EventTarget)) {
		if (T in e) pr(e);
		else if (!Array.isArray(e)) for (let t in e) {
			let n = e[t];
			typeof n == "object" && n && T in n && pr(n);
		}
	}
}
function pr(e, t = /* @__PURE__ */ new Set()) {
	if (typeof e == "object" && e && !(e instanceof EventTarget) && !t.has(e)) {
		t.add(e), e instanceof Date && e.getTime();
		for (let n in e) try {
			pr(e[n], t);
		} catch {}
		let n = p(e);
		if (n !== Object.prototype && n !== Array.prototype && n !== Map.prototype && n !== Set.prototype && n !== Date.prototype) {
			let t = u(n);
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
var mr = Symbol("events"), hr = /* @__PURE__ */ new Set(), gr = /* @__PURE__ */ new Set();
function _r(e, t, n, r = {}) {
	function i(e) {
		if (r.capture || Cr.call(t, e), !e.cancelBubble) return dt(() => n?.call(this, e));
	}
	return e.startsWith("pointer") || e.startsWith("touch") || e === "wheel" ? (i.__removed = !1, it(() => {
		i.__removed || t.addEventListener(e, i, r);
	})) : t.addEventListener(e, i, r), i;
}
function vr(e, t, n, r, i) {
	var a = {
		capture: r,
		passive: i
	}, o = _r(e, t, n, a);
	(t === document.body || t === window || t === document || t instanceof HTMLMediaElement) && Sn(() => {
		o.__removed = !0, t.removeEventListener(e, o, a);
	});
}
function yr(e, t, n) {
	(t[mr] ??= {})[e] = n;
}
function br(e) {
	for (var t = 0; t < e.length; t++) hr.add(e[t]);
	for (var n of gr) n(e);
}
var xr = null, Sr = !1;
function Cr(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	xr = e, Sr || (Sr = !0, setTimeout(() => {
		Sr = !1, xr = null;
	}));
	var o = 0, s = xr === e && e[mr];
	if (s) {
		var l = i.indexOf(s);
		if (l !== -1 && (t === document || t === window)) {
			e[mr] = t;
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
		var d = U, f = W;
		Gn(null), Kn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[mr]?.[r];
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
			e[mr] = t, delete e.currentTarget, Gn(d), Kn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var wr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function Tr(e) {
	return wr?.createHTML(e) ?? e;
}
function Er(e) {
	var t = mn("template");
	return t.innerHTML = Tr(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function Dr(e, t) {
	var n = W;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function J(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (E) return Dr(D, null), D;
		i === void 0 && (i = Er(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ L(i)));
		var t = r || an ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ L(t), s = t.lastChild;
			Dr(o, s);
		} else Dr(t, t);
		return t;
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Or(e, t, n = "svg") {
	var r = !e.startsWith("<!>"), i = !!(t & 1), a = `<${n}>${r ? e : "<!>" + e}</${n}>`, o;
	return () => {
		if (E) return Dr(D, null), D;
		if (!o) {
			var e = /* @__PURE__ */ L(Er(a));
			if (i) for (o = document.createDocumentFragment(); /* @__PURE__ */ L(e);) o.appendChild(/* @__PURE__ */ L(e));
			else o = /* @__PURE__ */ L(e);
		}
		var t = o.cloneNode(!0);
		if (i) {
			var n = /* @__PURE__ */ L(t), r = t.lastChild;
			Dr(n, r);
		} else Dr(t, t);
		return t;
	};
}
/*#__NO_SIDE_EFFECTS__*/
function kr(e, t) {
	return /* @__PURE__ */ Or(e, t, "svg");
}
function Ar() {
	if (E) return Dr(D, null), D;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = I();
	return e.append(t, n), Dr(t, n), e;
}
function Y(e, t) {
	if (E) {
		var n = W;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = D), Ee();
	} else e !== null && e.before(t);
}
// node_modules/svelte/src/utils.js
function jr(e) {
	return e.endsWith("capture") && e !== "gotpointercapture" && e !== "lostpointercapture";
}
var Mr = [
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
function Nr(e) {
	return Mr.includes(e);
}
var Pr = /* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split("."), Fr = {
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
function Ir(e) {
	return e = e.toLowerCase(), Fr[e] ?? e;
}
[...Pr];
var Lr = ["touchstart", "touchmove"];
function Rr(e) {
	return Lr.includes(e);
}
var zr = [
	"textarea",
	"script",
	"style",
	"title"
];
function Br(e) {
	return zr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function Vr(e) {
	let t = 0, n = Gt(0), r;
	return () => {
		xn() && (q(n), On(() => (t === 0 && (r = dr(() => e(() => Qt(n)))), t += 1, () => {
			it(() => {
				--t, t === 0 && (r?.(), r = void 0, Qt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var Hr = ie | ae;
function Ur(e, t, n, r) {
	new Wr(e, t, n, r);
}
var Wr = class {
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
	#m = Vr(() => (this.#p = Gt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = W;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = W.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = kn(() => {
			if (E) {
				let e = ke(this.#t);
				Ee();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, Hr), E && (this.#e = D);
	}
	#h() {
		try {
			this.#a = V(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		it(r), t && (this.#s = V(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? we() : (t = !0, n && We(), this.#s !== null && In(this.#s, () => {
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
					_n(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (te | re)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = V(() => e(this.#e)), it(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = I(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return V(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							_n(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(N)) : this.#u === 0 && (this.#e.before(e), this.#c = null, In(this.#o, () => {
					this.#o = null;
				}), this.#x(N));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = V(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				Bn(this.#a, e);
				let t = this.#n.pending;
				this.#o = V(() => t(this.#e));
			} else this.#x(N);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		st(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = W, n = U, r = A;
		Kn(this.#i), Gn(this.#i), Xe(this.#i.ctx);
		try {
			return Ft.ensure(), e();
		} finally {
			Kn(t), Gn(n), Xe(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && In(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, it(() => {
			this.#d = !1, this.#p && Yt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), q(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		N?.is_fork ? (this.#a && N.skip_effect(this.#a), this.#o && N.skip_effect(this.#o), this.#s && N.skip_effect(this.#s), N.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (H(this.#a), null), this.#o &&= (H(this.#o), null), this.#s &&= (H(this.#s), null), E && (O(this.#t), De(), O(Oe()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return V(() => {
						var r = W;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return _n(e, this.#i.parent), null;
				}
			}));
		};
		it(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					_n(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => _n(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function Gr(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[_e] ??= e.nodeValue) && (e[_e] = n, e.nodeValue = `${n}`);
}
function Kr(e, t) {
	return Jr(e, t);
}
var qr = /* @__PURE__ */ new Map();
function Jr(t, { target: n, anchor: r, props: i = {}, events: a, context: o, intro: c = !0, transformError: l }) {
	cn();
	var u = void 0, d = Tn(() => {
		var c = r ?? n.appendChild(I());
		Ur(c, { pending: () => {} }, (n) => {
			Qe({});
			var r = A;
			if (o && (r.c = o), a && (i.$$events = a), E && Dr(n, null), u = t(n, i) || et(), E && (W.nodes.end = D, D === null || D.nodeType !== 8 || D.data !== "]")) throw Se(), e;
			$e();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = Rr(r);
					for (let e of [n, document]) {
						var a = qr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), qr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, Cr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(s(hr)), gr.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = qr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, Cr), t.delete(e), t.size === 0 && qr.delete(r)) : t.set(e, i);
			}
			gr.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return Yr.set(u, d), u;
}
var Yr = /* @__PURE__ */ new WeakMap();
function Xr(e, t) {
	let n = Yr.get(e);
	return n ? (Yr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var Zr = class {
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
			if (n) Rn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (Rn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (H(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						Bn(r, t), t.append(I()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else H(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), In(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (H(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = N, r = pn();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = I();
				i.append(a), this.#n.set(e, {
					effect: V(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, V(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else E && (this.anchor = D), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/snippet.js
function Qr(e, t, ...n) {
	var r = new Zr(e);
	kn(() => {
		let e = t() ?? null;
		r.ensure(e, e && ((t) => e(t, ...n)));
	}, ie);
}
function $r(e) {
	A === null && Ne("onMount"), Cn(() => {
		let t = dr(e);
		if (typeof t == "function") return t;
	});
}
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function X(e, t, n = !1) {
	var r;
	E && (r = D, Ee());
	var i = new Zr(e), a = n ? ie : 0;
	function o(e, t) {
		if (E) {
			var n = ke(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Oe();
				O(a), i.anchor = a, Te(!1), i.ensure(e, t), Te(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	kn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/key.js
var ei = Symbol("NaN");
function ti(e, t, n) {
	E && Ee();
	var r = new Zr(e), i = !tt();
	kn(() => {
		var e = t();
		e !== e && (e = ei), i && typeof e == "object" && e && (e = {}), r.ensure(e, n);
	});
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function ni(e, t) {
	return t;
}
function ri(e, t, n) {
	for (var r = [], i = t.length, a, o = t.length, c = 0; c < i; c++) {
		let n = t[c];
		In(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					ii(e, s(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --o;
		}, !1);
	}
	if (o === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			fn(d), d.append(u), e.items.clear();
		}
		ii(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function ii(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= se, Bn(a, document.createDocumentFragment())) : H(t[i], n);
	}
}
var ai;
function oi(e, t, n, r, a, o = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = E ? O(/* @__PURE__ */ L(u)) : u.appendChild(I());
	}
	E && Ee();
	var d = null, f = /* @__PURE__ */ yt(() => {
		var e = n();
		return i(e) ? e : e == null ? [] : s(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = q(f);
			_.fallback = d, ci(_, n, c, t, r), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= se, ui(d, null, c)) : Rn(d) : In(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: kn(() => {
			var e = q(f), i = e.length;
			let s = !1;
			E && ke(c) === "[!" != (i === 0) && (c = Oe(), O(c), Te(!1), s = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = N, v = pn(), y = 0; y < i; y += 1) {
				E && D.nodeType === 8 && D.data === "]" && (c = D, s = !0, Te(!1));
				var b = e[y], x = r(b, y), S = m ? null : l.get(x);
				S ? (S.v && Yt(S.v, b), S.i && Yt(S.i, y), v && _.unskip_effect(S.e)) : (S = li(l, m ? c : ai ??= I(), b, x, y, a, t, n), m || (S.e.f |= se), l.set(x, S)), u.add(x);
			}
			if (i === 0 && o && !d && (m ? d = V(() => o(c)) : (d = V(() => o(ai ??= I())), d.f |= se)), i > u.size && Fe("", "", ""), E && i > 0 && O(Oe()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			s && Te(!0), q(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, E && (c = D);
}
function si(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function ci(e, t, n, r, i) {
	var a = !!(r & 8), o = t.length, c = e.items, l = si(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < o; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < o; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (Rn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= se, _ === l) ui(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), di(e, d, _), di(e, _, y), ui(_, y, n), d = _, p = [], m = [], l = si(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) ui(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					di(e, S.prev, C.next), di(e, d, S), di(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), ui(_, l, n), di(e, _.prev, _.next), di(e, _, d === null ? e.effect.first : d.next), di(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = si(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = si(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (ii(e, s(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var w = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || w.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && w.push(l), l = si(l.next);
		var ee = w.length;
		if (ee > 0) {
			var te = r & 4 && o === 0 ? n : null;
			if (a) {
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.measure();
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.fix();
			}
			ri(e, w, te);
		}
	}
	a && it(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function li(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Gt(n) : /* @__PURE__ */ Kt(n, !1, !1) : null, l = o & 2 ? Gt(i) : null;
	return {
		v: c,
		i: l,
		e: V(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function ui(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ ln(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function di(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/client/dom/blocks/svelte-component.js
function fi(e, t, n) {
	var r;
	E && (r = D, Ee());
	var i = new Zr(e);
	kn(() => {
		var e = t() ?? null;
		if (E && ke(r) === "[" != (e !== null)) {
			var a = Oe();
			O(a), i.anchor = a, Te(!1), i.ensure(e, e && ((t) => n(t, e))), Te(!0);
		} else i.ensure(e, e && ((t) => n(t, e)));
	}, ie);
}
// node_modules/svelte/src/internal/client/dom/blocks/svelte-element.js
function pi(e, t, n, i, a, o) {
	let s = E;
	E && Ee();
	var c = null;
	E && D.nodeType === 1 && (c = D, Ee());
	var l = E ? D : e, u = new Zr(l, !1);
	kn(() => {
		let e = t() || null;
		var o = a ? a() : n || e === "svg" ? r : void 0;
		if (e === null) u.ensure(null, null);
		else return u.ensure(e, (t) => {
			if (e) {
				var n = E;
				if (c = E ? c : mn(e, o), Dr(c, c), i) {
					var r = null;
					E && Br(e) && c.append(r = document.createComment(""));
					var a = E ? /* @__PURE__ */ L(c) : c.appendChild(I());
					E && (a === null ? Te(!1) : O(a)), i(c, a), r?.remove();
				}
				W.nodes.end = c, n || t.before(c);
			}
			E && O(t);
		}), () => {};
	}, ie), Sn(() => {}), s && (Te(!0), O(l));
}
// node_modules/svelte/src/internal/client/dom/elements/actions.js
function mi(e, t, n) {
	En(() => {
		var r = dr(() => t(e, n?.()) || {});
		if (n && r?.update) {
			var i = !1, a = {};
			On(() => {
				var e = n();
				fr(e), i && je(a, e) && (a = e, r.update(e));
			}), i = !0;
		}
		if (r?.destroy) return () => r.destroy();
	});
}
// node_modules/svelte/src/internal/client/dom/elements/attachments.js
function hi(e, t) {
	var n = void 0, r;
	An(() => {
		n !== (n = t()) && (r &&= (H(r), null), n && (r = V(() => {
			En(() => n(e));
		})));
	});
}
// node_modules/clsx/dist/clsx.mjs
function gi(e) {
	var t, n, r = "";
	if (typeof e == "string" || typeof e == "number") r += e;
	else if (typeof e == "object") {
		if (Array.isArray(e)) {
			var i = e.length;
			for (t = 0; t < i; t++) e[t] && (n = gi(e[t])) && (r && (r += " "), r += n);
		} else for (n in e) e[n] && (r && (r += " "), r += n);
	}
	return r;
}
function _i() {
	for (var e, t, n = 0, r = "", i = arguments.length; n < i; n++) (e = arguments[n]) && (t = gi(e)) && (r && (r += " "), r += t);
	return r;
}
// node_modules/svelte/src/internal/shared/attributes.js
function vi(e) {
	return typeof e == "object" ? _i(e) : e ?? "";
}
var yi = [..." 	\n\r\f\xA0\v﻿"];
function bi(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || yi.includes(r[o - 1])) && (s === r.length || yi.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
function xi(e, t = !1) {
	var n = t ? " !important;" : ";", r = "";
	for (var i of Object.keys(e)) {
		var a = e[i];
		a != null && a !== "" && (r += " " + i + ": " + a + n);
	}
	return r;
}
function Si(e) {
	return e[0] !== "-" || e[1] !== "-" ? e.toLowerCase() : e;
}
function Ci(e, t) {
	if (t) {
		var n = "", r, i;
		if (Array.isArray(t) ? (r = t[0], i = t[1]) : r = t, e) {
			e = String(e).replaceAll(/\/\*.*?\*\//g, "").trim();
			var a = !1, o = 0, s = !1, c = [];
			r && c.push(...Object.keys(r).map(Si)), i && c.push(...Object.keys(i).map(Si));
			var l = 0, u = -1;
			let t = e.length;
			for (var d = 0; d < t; d++) {
				var f = e[d];
				if (s ? f === "/" && e[d - 1] === "*" && (s = !1) : a ? a === f && (a = !1) : f === "/" && e[d + 1] === "*" ? s = !0 : f === "\"" || f === "'" ? a = f : f === "(" ? o++ : f === ")" && o--, !s && a === !1 && o === 0) {
					if (f === ":" && u === -1) u = d;
					else if (f === ";" || d === t - 1) {
						if (u !== -1) {
							var p = Si(e.substring(l, u).trim());
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
		return r && (n += xi(r)), i && (n += xi(i, !0)), n = n.trim(), n === "" ? null : n;
	}
	return e == null ? null : String(e);
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function wi(e, t, n, r, i, a) {
	var o = e[he];
	if (E || o !== n || o === void 0) {
		var s = bi(n, r, a);
		E && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[he] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/style.js
function Ti(e, t = {}, n, r) {
	for (var i in n) {
		var a = n[i];
		t[i] !== a && (n[i] == null ? e.style.removeProperty(i) : e.style.setProperty(i, a, r));
	}
}
function Ei(e, t, n, r) {
	var i = e[ge];
	if (E || i !== t) {
		var a = Ci(t, r);
		E && a === e.getAttribute("style") || (a == null ? e.removeAttribute("style") : e.style.cssText = a), e[ge] = t;
	} else r && (Array.isArray(r) ? (Ti(e, n?.[0], r[0]), Ti(e, n?.[1], r[1], "important")) : Ti(e, n, r));
	return r;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/select.js
function Di(e, t) {
	t ? e.hasAttribute("selected") || e.setAttribute("selected", "") : e.removeAttribute("selected");
}
function Oi(e, t) {
	var n = !("__defaultValue" in e);
	(n || e.__defaultValue !== t) && (e.__defaultValue = t, ki(e, !n || "__value" in e));
}
function ki(e, t) {
	var n = e.__defaultValue, r = e.multiple, a = r ? n ?? [] : null;
	if (!r || i(a)) {
		var o = e.selectedIndex, s = t && r ? new Set(e.selectedOptions) : null;
		for (var c of e.options) {
			var l = Mi(c);
			Di(c, r ? a.includes(l) : nn(l, n));
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
function Ai(e, t, n = !1) {
	if (e.multiple) {
		if (t == null) return;
		if (!i(t)) return Ce();
		for (var r of e.options) r.selected = t.includes(Mi(r));
	} else {
		for (r of e.options) if (nn(Mi(r), t)) {
			r.selected = !0;
			return;
		}
		n && t === void 0 || (e.selectedIndex = -1);
	}
}
function ji(e) {
	var t = new MutationObserver((t) => {
		t.every(Ni) || ("__defaultValue" in e && ki(e, !1), "__value" in e && Ai(e, e.__value));
	});
	t.observe(e, {
		childList: !0,
		subtree: !0,
		attributes: !0,
		attributeFilter: ["value"]
	}), Sn(() => {
		t.disconnect();
	});
}
function Mi(e) {
	return "__value" in e ? e.__value : e.value;
}
function Ni(e) {
	if (e.target.closest("selectedcontent") !== null) return !0;
	if (e.type === "childList") {
		var t = [...e.addedNodes, ...e.removedNodes];
		return t.length > 0 && t.every((e) => e.nodeName === "SELECTEDCONTENT");
	}
	return !1;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Pi = Symbol("class"), Fi = Symbol("style"), Ii = Symbol("is custom element"), Li = Symbol("is html"), Ri = be ? "link" : "LINK", zi = be ? "input" : "INPUT", Bi = be ? "option" : "OPTION", Vi = be ? "select" : "SELECT", Hi = be ? "progress" : "PROGRESS";
function Ui(e) {
	if (E) {
		var t = !1, n = () => {
			if (!t) {
				if (t = !0, e.hasAttribute("value")) {
					var n = e.value;
					Z(e, "value", null), e.value = n;
				}
				if (e.hasAttribute("checked")) {
					var r = e.checked;
					Z(e, "checked", null), e.checked = r;
				}
			}
		};
		e[ve] = n, it(n), ut();
	}
}
function Z(e, t, n, r) {
	var i = Ki(e);
	if (!(E && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Ri))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[pe] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Ji(e).has(t) ? e[t] = n : (!E || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Wi(e, n, r, i, a = !1, o = !1) {
	E && a && e.nodeName === zi && ("defaultValue" in r || "defaultChecked" in r || Ui(e));
	var s = Ki(e), c = s[Ii], l = !s[Li];
	let u = E && c;
	u && Te(!1);
	var d = n || {}, f = e.nodeName === Bi, p = e.nodeName === Vi;
	try {
		for (var m in n) !(m in r) && m[0] + m[1] !== "$$" && (r[m] = null);
		r.class ? r.class = vi(r.class) : (i || r[Pi]) && (r.class = null), r[Fi] && (r.style ??= null);
		var h = Ji(e);
		if (e.nodeName === zi && "type" in r && ("value" in r || "__value" in r)) {
			var g = r.type;
			(g !== d.type || g === void 0 && e.hasAttribute("type")) && (d.type = g, Z(e, "type", g, o));
		}
		for (let a in r) {
			let u = r[a];
			if (f && a === "value" && u == null) e.value = e.__value = "", d[a] = u;
			else if (a === "class") wi(e, e.namespaceURI === "http://www.w3.org/1999/xhtml", u, i, n?.[Pi], r[Pi]), d[a] = u, d[Pi] = r[Pi];
			else if (a === "style") Ei(e, u, n?.[Fi], r[Fi]), d[a] = u, d[Fi] = r[Fi];
			else {
				var _ = d[a];
				if (u !== _ || u === void 0 && e.hasAttribute(a)) {
					d[a] = u;
					var v = a[0] + a[1];
					if (v !== "$$") {
						if (v === "on") {
							let t = {}, n = "$$" + a, r = a.slice(2);
							var y = Nr(r);
							if (jr(r) && (r = r.slice(0, -7), t.capture = !0), !y && _) {
								if (u != null) continue;
								e.removeEventListener(r, d[n], t), d[n] = null;
							}
							if (y) yr(r, e, u), br([r]);
							else if (u != null) {
								function i(e) {
									d[a].call(this, e);
								}
								d[n] = _r(r, e, i, t);
							}
						} else if (a === "style") Z(e, a, u);
						else if (a === "autofocus") ct(e, !!u);
						else if (!c && (a === "__value" || a === "value" && u != null)) e.__value = u, (_ == null || e.value !== u || u === 0 && e.nodeName === Hi) && (e.value = u);
						else if (a === "selected" && f) Di(e, u);
						else {
							var b = a;
							l || (b = Ir(b));
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
							} else x || (c || typeof u != "string") && h.has(b) ? (e[b] = u, b in s && (s[b] = t)) : typeof u != "function" && Z(e, b, u, o);
						}
					}
				}
			}
		}
	} finally {
		u && Te(!0);
	}
	return d;
}
function Gi(e, t, n = [], r = [], i = [], a, o = !1, s = !1) {
	ft(i, n, r, (n) => {
		var r = void 0, i = {}, c = e.nodeName === Vi, l = !1;
		if (An(() => {
			var u = t(...n.map(q)), d = Wi(e, r, u, a, o, s);
			if (l && c) {
				var f = e;
				"defaultValue" in u && Oi(f, u.defaultValue), "value" in u && Ai(f, u.value);
			}
			for (let e of Object.getOwnPropertySymbols(i)) u[e] || H(i[e]);
			for (let t of Object.getOwnPropertySymbols(u)) {
				var p = u[t];
				t.description !== "@attach" || r && p === r[t] || (i[t] && H(i[t]), i[t] = V(() => hi(e, () => p))), d[t] = p;
			}
			r = d;
		}), c) {
			var u = e;
			En(() => {
				var e = r;
				"defaultValue" in e && Oi(u, e.defaultValue), Ai(u, e.value, !0), ji(u);
			});
		}
		l = !0;
	});
}
function Ki(e) {
	return e[me] ??= {
		[Ii]: e.nodeName.includes("-"),
		[Li]: e.namespaceURI === n
	};
}
var qi = /* @__PURE__ */ new Map();
function Ji(e) {
	var t = e.getAttribute("is") || e.nodeName, n = qi.get(t);
	if (n) return n;
	qi.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = u(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = p(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/this.js
function Yi(e, t) {
	return e === t || e?.[T] === t;
}
function Xi(e = et(), t, n, r) {
	var i = A.r, a = W;
	return En(() => {
		var o, s;
		return On(() => {
			o = s, s = r?.() || [], dr(() => {
				Yi(n(...s), e) || (t(e, ...s), o && Yi(n(...o), e) && t(null, ...o));
			});
		}), () => {
			let r = a;
			for (; r !== i && r.parent !== null && r.parent.f & 33554432;) r = r.parent;
			let o = () => {
				s && Yi(n(...s), e) && t(null, ...s);
			}, c = r.teardown;
			r.teardown = () => {
				o(), c?.();
			};
		};
	}), e;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var Zi = !1;
function Qi(e) {
	var t = Zi;
	try {
		return Zi = !1, [e(), Zi];
	} finally {
		Zi = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
var $i = {
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
function Q(e, t, n) {
	return new Proxy({
		props: e,
		exclude: t
	}, $i);
}
var ea = {
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
		if (t === T || t === fe) return !1;
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
function ta(...e) {
	return new Proxy({ props: e }, ea);
}
function na(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, c = !0, u = void 0, d = () => o && i ? (u ??= /* @__PURE__ */ gt(r), q(u)) : (c && (c = !1, s = o ? dr(r) : r), s);
	let f;
	if (a) {
		var p = T in e || fe in e;
		f = l(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = Qi(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Be(t), f(m)));
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
	var v = !1, y = (n & 1 ? gt : yt)(() => (v = !1, g()));
	a && q(y);
	var b = W;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? q(y) : i && a ? en(e) : e;
			return F(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return Hn && v || b.f & 16384 ? y.v : q(y);
	});
}
// kit/plugin.svelte.js
function ra(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href, r = null, i = !1, a = () => (r ??= new Promise((e) => {
		let t = document.querySelector(`link[href="${n}"]`) ?? document.createElement("link");
		t.sheet ? e() : (t.addEventListener("load", () => e(), { once: !0 }), t.addEventListener("error", () => e(), { once: !0 }), t.isConnected || (t.rel = "stylesheet", t.href = n, document.head.append(t)));
	}).then(() => {
		i = !0;
	}), r);
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ P(en({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return q(this.#e);
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
			this.style.display = "contents";
			let e = a();
			this.app ||= Kr(t, {
				target: this,
				props: this.props
			}), i || (this.style.visibility = "hidden", e.then(() => {
				this.style.visibility = "";
			}));
		}
		disconnectedCallback() {
			this.app && Xr(this.app), this.app = null;
		}
	});
}
function ia(e, t) {
	let n = en({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			Zt(r);
		}
	}), r = /* @__PURE__ */ P(0);
	Cn(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, Zt(r);
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
	return Cn(() => {
		let n = e(), c = t();
		q(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
function aa(e) {
	let t = en({ ...e()?.params?.() ?? {} });
	return Cn(() => {
		let n = e();
		if (!n) return;
		let r = (e) => {
			for (let n of Object.keys(t)) n in e || delete t[n];
			Object.assign(t, e);
		};
		return r(n.params()), n.onParams(r);
	}), t;
}
function oa(e, t) {
	let n = t.api.mount(e, t.slot, t.target, t.context);
	return {
		update(r) {
			n && JSON.stringify(r.target) === JSON.stringify(t.target) ? n.context(r.context) : (n?.destroy(), n = r.api.mount(e, r.slot, r.target, r.context)), t = r;
		},
		destroy() {
			n?.destroy();
		}
	};
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// node_modules/@lucide/svelte/dist/utils/defaultAttributes.js
var sa = {
	xmlns: "http://www.w3.org/2000/svg",
	width: 24,
	height: 24,
	viewBox: "0 0 24 24",
	fill: "none",
	stroke: "currentColor",
	"stroke-width": 2,
	"stroke-linecap": "round",
	"stroke-linejoin": "round"
}, ca = (...e) => e.filter((e, t, n) => !!e && e.trim() !== "" && n.indexOf(e) === t).join(" ").trim();
// node_modules/@lucide/svelte/dist/utils/buildLucideIconNode.js
function la(e) {
	return e != null;
}
function ua(e, t = {}) {
	let n = t.attributeNames ?? {}, r = (e) => n[e] ?? e, i = e.size ?? e.width ?? sa.width, a = e.size ?? e.height ?? sa.height, o = e.aliases?.filter((e) => typeof e == "string" && e.trim() !== "").map((e) => `lucide-${e}`) ?? [], s = [...e.name ? [`lucide-${e.name}`] : [], ...o], c = t.className?.split(" ").filter(Boolean) ?? [], l = t.includeDefaultClasses === !1 ? ca(...c) : ca("lucide", ...s, ...c), u = t.absoluteStrokeWidth ? Number(t.strokeWidth ?? sa["stroke-width"]) * Number(e.size ?? e.width ?? sa.width) / Number(t.size ?? t.width ?? sa.width) : t.strokeWidth ?? sa["stroke-width"];
	return [
		"svg",
		{
			...Object.entries(sa).reduce((e, [t, n]) => (e[r(t)] = n, e), {}),
			..."color" in t && t.color && { [r("stroke")]: t.color },
			..."size" in t && la(t.size) && {
				[r("width")]: t.size,
				[r("height")]: t.size
			},
			..."width" in t && la(t.width) && { [r("width")]: t.width },
			..."height" in t && la(t.height) && { [r("height")]: t.height },
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
var da = (e) => {
	for (let t in e) if (t.startsWith("aria-") || t === "role" || t === "title") return !0;
	return !1;
}, fa = Symbol("lucide-context"), pa = () => Ze(fa), ma = /* @__PURE__ */ new Set([
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
]), ha = /* @__PURE__ */ kr("<svg><!><!></svg>");
function ga(e, t) {
	Qe(t, !0);
	let n = pa() ?? {}, r = na(t, "color", 19, () => n.color ?? "currentColor"), i = na(t, "size", 19, () => n.size ?? 24), a = na(t, "width", 19, i), o = na(t, "height", 19, i), s = na(t, "strokeWidth", 19, () => n.strokeWidth ?? 2), c = na(t, "absoluteStrokeWidth", 19, () => n.absoluteStrokeWidth ?? !1), l = na(t, "nonScalingStroke", 19, () => n.nonScalingStroke ?? !1), u = na(t, "iconNode", 19, () => []), d = na(t, "icon", 19, () => ({
		node: u(),
		aliases: [],
		size: 24
	})), f = /* @__PURE__ */ Q(t, ma), p = /* @__PURE__ */ M(() => !!t.children || da(f)), m = /* @__PURE__ */ M(() => ua(d(), {
		color: r(),
		width: a(),
		height: o(),
		strokeWidth: s(),
		absoluteStrokeWidth: c(),
		nonScalingStroke: l(),
		className: ca("lucide-icon", n.class),
		hasA11yProp: q(p),
		attributes: f
	})), h = /* @__PURE__ */ M(() => b(q(m), 3)), _ = /* @__PURE__ */ M(() => q(h)[1]), v = /* @__PURE__ */ M(() => y(q(h)[2], () => [], !0)), x = /* @__PURE__ */ M(() => ({
		...q(_),
		class: [...q(_).class.split(" "), t.class]
	}));
	var S = ha();
	Gi(S, () => ({ ...q(x) }));
	var C = R(S);
	oi(C, 17, () => q(v), ni, (e, t) => {
		var n = /* @__PURE__ */ M(() => b(q(t), 2));
		let r = () => q(n)[0], i = () => q(n)[1];
		var a = Ar();
		pi(un(a), r, !0, (e, t) => {
			Gi(e, () => ({ ...i() }));
		}), Y(e, a);
	}), Qr(z(C), () => t.children ?? g), k(S), Y(e, S), $e();
}
// node_modules/@lucide/svelte/dist/icons/inbox.svelte
var _a = {
	name: "inbox",
	size: 24,
	node: [["polyline", { points: "22 12 16 12 14 15 10 15 8 12 2 12" }], ["path", { d: "M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z" }]]
};
_a.node;
var va = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function ya(e, t) {
	let n = /* @__PURE__ */ Q(t, va);
	ga(e, ta(() => n, { get icon() {
		return _a;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/message-square.svelte
var ba = {
	name: "message-square",
	size: 24,
	node: [["path", { d: "M22 17a2 2 0 0 1-2 2H6.828a2 2 0 0 0-1.414.586l-2.202 2.202A.71.71 0 0 1 2 21.286V5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2z" }]]
};
ba.node;
var xa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Sa(e, t) {
	let n = /* @__PURE__ */ Q(t, xa);
	ga(e, ta(() => n, { get icon() {
		return ba;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/list-checks.svelte
var Ca = {
	name: "list-checks",
	size: 24,
	node: [
		["path", { d: "M13 5h8" }],
		["path", { d: "M13 12h8" }],
		["path", { d: "M13 19h8" }],
		["path", { d: "m3 17 2 2 4-4" }],
		["path", { d: "m3 7 2 2 4-4" }]
	]
};
Ca.node;
var wa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ta(e, t) {
	let n = /* @__PURE__ */ Q(t, wa);
	ga(e, ta(() => n, { get icon() {
		return Ca;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/activity.svelte
var Ea = {
	name: "activity",
	size: 24,
	node: [["path", { d: "M22 12h-2.48a2 2 0 0 0-1.93 1.46l-2.35 8.36a.25.25 0 0 1-.48 0L9.24 2.18a.25.25 0 0 0-.48 0l-2.35 8.36A2 2 0 0 1 4.49 12H2" }]]
};
Ea.node;
var Da = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Oa(e, t) {
	let n = /* @__PURE__ */ Q(t, Da);
	ga(e, ta(() => n, { get icon() {
		return Ea;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/terminal.svelte
var ka = {
	name: "terminal",
	size: 24,
	node: [["path", { d: "M12 19h8" }], ["path", { d: "m4 17 6-6-6-6" }]]
};
ka.node;
var Aa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function ja(e, t) {
	let n = /* @__PURE__ */ Q(t, Aa);
	ga(e, ta(() => n, { get icon() {
		return ka;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/boxes.svelte
var Ma = {
	name: "boxes",
	size: 24,
	node: [
		["path", { d: "M2.97 12.92A2 2 0 0 0 2 14.63v3.24a2 2 0 0 0 .97 1.71l3 1.8a2 2 0 0 0 2.06 0L12 19v-5.5l-5-3-4.03 2.42Z" }],
		["path", { d: "m7 16.5-4.74-2.85" }],
		["path", { d: "m7 16.5 5-3" }],
		["path", { d: "M7 16.5v5.17" }],
		["path", { d: "M12 13.5V19l3.97 2.38a2 2 0 0 0 2.06 0l3-1.8a2 2 0 0 0 .97-1.71v-3.24a2 2 0 0 0-.97-1.71L17 10.5l-5 3Z" }],
		["path", { d: "m17 16.5-5-3" }],
		["path", { d: "m17 16.5 4.74-2.85" }],
		["path", { d: "M17 16.5v5.17" }],
		["path", { d: "M7.97 4.42A2 2 0 0 0 7 6.13v4.37l5 3 5-3V6.13a2 2 0 0 0-.97-1.71l-3-1.8a2 2 0 0 0-2.06 0l-3 1.8Z" }],
		["path", { d: "M12 8 7.26 5.15" }],
		["path", { d: "m12 8 4.74-2.85" }],
		["path", { d: "M12 13.5V8" }]
	]
};
Ma.node;
var Na = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Pa(e, t) {
	let n = /* @__PURE__ */ Q(t, Na);
	ga(e, ta(() => n, { get icon() {
		return Ma;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/rocket.svelte
var Fa = {
	name: "rocket",
	size: 24,
	node: [
		["path", { d: "M12 15v5s3.03-.55 4-2c1.08-1.62 0-5 0-5" }],
		["path", { d: "M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09" }],
		["path", { d: "M9 12a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.4 22.4 0 0 1-4 2z" }],
		["path", { d: "M9 12H4s.55-3.03 2-4c1.62-1.08 5 .05 5 .05" }]
	]
};
Fa.node;
var Ia = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function La(e, t) {
	let n = /* @__PURE__ */ Q(t, Ia);
	ga(e, ta(() => n, { get icon() {
		return Fa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/sun.svelte
var Ra = {
	name: "sun",
	size: 24,
	node: [
		["circle", {
			cx: "12",
			cy: "12",
			r: "4"
		}],
		["path", { d: "M12 2v2" }],
		["path", { d: "M12 20v2" }],
		["path", { d: "m4.93 4.93 1.41 1.41" }],
		["path", { d: "m17.66 17.66 1.41 1.41" }],
		["path", { d: "M2 12h2" }],
		["path", { d: "M20 12h2" }],
		["path", { d: "m6.34 17.66-1.41 1.41" }],
		["path", { d: "m19.07 4.93-1.41 1.41" }]
	]
};
Ra.node;
var za = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ba(e, t) {
	let n = /* @__PURE__ */ Q(t, za);
	ga(e, ta(() => n, { get icon() {
		return Ra;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/moon.svelte
var Va = {
	name: "moon",
	size: 24,
	node: [["path", { d: "M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401" }]]
};
Va.node;
var Ha = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ua(e, t) {
	let n = /* @__PURE__ */ Q(t, Ha);
	ga(e, ta(() => n, { get icon() {
		return Va;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/puzzle.svelte
var Wa = {
	name: "puzzle",
	size: 24,
	node: [["path", { d: "M15.39 4.39a1 1 0 0 0 1.68-.474 2.5 2.5 0 1 1 3.014 3.015 1 1 0 0 0-.474 1.68l1.683 1.682a2.414 2.414 0 0 1 0 3.414L19.61 15.39a1 1 0 0 1-1.68-.474 2.5 2.5 0 1 0-3.014 3.015 1 1 0 0 1 .474 1.68l-1.683 1.682a2.414 2.414 0 0 1-3.414 0L8.61 19.61a1 1 0 0 0-1.68.474 2.5 2.5 0 1 1-3.014-3.015 1 1 0 0 0 .474-1.68l-1.683-1.682a2.414 2.414 0 0 1 0-3.414L4.39 8.61a1 1 0 0 1 1.68.474 2.5 2.5 0 1 0 3.014-3.015 1 1 0 0 1-.474-1.68l1.683-1.682a2.414 2.414 0 0 1 3.414 0z" }]]
};
Wa.node;
var Ga = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ka(e, t) {
	let n = /* @__PURE__ */ Q(t, Ga);
	ga(e, ta(() => n, { get icon() {
		return Wa;
	} }));
}
// plugins/desk/panel.svelte.js
var qa = 160, Ja = 120, Ya = 360, Xa = 12;
function Za(e) {
	return Math.round(Math.max(qa, Math.min(window.innerHeight - Ja, e)));
}
var Qa = Number(localStorage.getItem("desk-console-h")), $a = window.matchMedia("(max-width: 768px)"), eo = en({
	h: Za(Qa > 0 ? Qa : Ya),
	mobile: $a.matches
});
$a.addEventListener("change", (e) => {
	eo.mobile = e.matches;
}), window.addEventListener("resize", () => {
	eo.h = Za(eo.h);
});
function to(e) {
	eo.h = Za(e), localStorage.setItem("desk-console-h", String(eo.h));
}
function no() {
	to(Ya);
}
function ro() {
	try {
		let e = JSON.parse(localStorage.getItem("desk-console-tabs") || "[]");
		return Array.isArray(e) ? e.filter((e) => e && typeof e.id == "string").slice(0, Xa) : [];
	} catch {
		return [];
	}
}
var $ = en({ list: ro() });
function io() {
	localStorage.setItem("desk-console-tabs", JSON.stringify($.list));
}
function ao(e) {
	if (!e.session) return;
	let t = {
		id: e.session,
		item: e.item || "",
		ro: e.ro || ""
	}, n = $.list.findIndex((e) => e.id === t.id);
	if (n < 0) $.list = [...$.list, t].slice(-12);
	else if ($.list[n].item !== t.item || $.list[n].ro !== t.ro) $.list[n] = t;
	else return;
	io();
}
function oo(e, t) {
	e.openSession(t.id, {
		item: t.item,
		readonly: t.ro === "1"
	});
}
function so(e, t, n) {
	let r = $.list.findIndex((e) => e.id === n);
	if (r < 0 || ($.list = $.list.filter((e) => e.id !== n), io(), t !== n)) return;
	let i = $.list[Math.min(r, $.list.length - 1)];
	i ? oo(e, i) : e.go({
		session: "",
		item: "",
		ro: ""
	});
}
// node_modules/@lucide/svelte/dist/icons/chevron-down.svelte
var co = {
	name: "chevron-down",
	size: 24,
	node: [["path", { d: "m6 9 6 6 6-6" }]]
};
co.node;
var lo = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function uo(e, t) {
	let n = /* @__PURE__ */ Q(t, lo);
	ga(e, ta(() => n, { get icon() {
		return co;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/chevron-up.svelte
var fo = {
	name: "chevron-up",
	size: 24,
	node: [["path", { d: "m18 15-6-6-6 6" }]]
};
fo.node;
var po = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function mo(e, t) {
	let n = /* @__PURE__ */ Q(t, po);
	ga(e, ta(() => n, { get icon() {
		return fo;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/x.svelte
var ho = {
	name: "x",
	size: 24,
	node: [["path", { d: "M18 6 6 18" }], ["path", { d: "m6 6 12 12" }]]
};
ho.node;
var go = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function _o(e, t) {
	let n = /* @__PURE__ */ Q(t, go);
	ga(e, ta(() => n, { get icon() {
		return ho;
	} }));
}
// plugins/desk/Console.svelte
var vo = /* @__PURE__ */ J("<div class=\"grip svelte-qk9xya\" role=\"slider\" aria-orientation=\"vertical\" aria-label=\"console height\" aria-valuemin=\"160\" tabindex=\"0\"></div>"), yo = /* @__PURE__ */ J("<div role=\"tab\" data-testid=\"console-tab\"><button type=\"button\" class=\"tl svelte-qk9xya\"> </button> <button type=\"button\" class=\"tx svelte-qk9xya\" data-testid=\"console-tab-close\"><!></button></div>"), bo = /* @__PURE__ */ J("<button type=\"button\" class=\"ic svelte-qk9xya\" aria-label=\"Minimize\" title=\"Minimize (Esc)\" data-testid=\"session-panel-close\"><!></button>"), xo = /* @__PURE__ */ J("<button type=\"button\" class=\"ic svelte-qk9xya\" aria-label=\"Restore\" title=\"Restore\" data-testid=\"console-restore\"><!></button>"), So = /* @__PURE__ */ J("<div class=\"plugin-host\"></div>"), Co = /* @__PURE__ */ J("<p class=\"faint none svelte-qk9xya\">No plugin shows sessions (slot <span class=\"mono\">session</span>).</p>"), wo = /* @__PURE__ */ J("<div class=\"body svelte-qk9xya\"><!></div>"), To = /* @__PURE__ */ J("<section tabindex=\"-1\" aria-label=\"sessions console\"><!> <div class=\"bar svelte-qk9xya\"><div class=\"tabs svelte-qk9xya\" role=\"tablist\"></div> <!></div> <!></section>");
function Eo(e, t) {
	Qe(t, !0);
	let n = /* @__PURE__ */ M(() => t.url.session || ""), r = /* @__PURE__ */ M(() => t.url.item || ""), i = /* @__PURE__ */ M(() => t.url.ro === "1");
	Cn(() => {
		ao({
			session: q(n),
			item: q(r),
			ro: t.url.ro || ""
		});
	});
	let a = /* @__PURE__ */ M(() => t.api.contributions("session").at(-1)), o = ia(() => t.api, () => $.list.length ? "api/sessions?since=720h" : null), s = /* @__PURE__ */ M(() => Object.fromEntries((o.data ?? []).map((e) => [e.id, e]))), c = (e) => {
		let t = q(s)[e];
		return t?.ticket?.pool || t?.display_name || e;
	}, l = (e) => {
		let t = q(s)[e];
		return [
			t?.display_name || t?.ticket?.title,
			t?.model,
			e
		].filter(Boolean).join("\n");
	};
	Cn(() => {
		if (!eo.mobile || !q(n)) return;
		let e = document.body.style.overflow;
		return document.body.style.overflow = "hidden", () => {
			document.body.style.overflow = e;
		};
	});
	let u = /* @__PURE__ */ P(void 0);
	Cn(() => {
		q(n) && q(u)?.focus({ preventScroll: !0 });
	});
	function d(e) {
		let t = e.currentTarget;
		t.setPointerCapture(e.pointerId);
		let n = q(u).getBoundingClientRect().bottom, r = (e) => to(n - e.clientY), i = () => {
			t.removeEventListener("pointermove", r), t.removeEventListener("pointerup", i), t.removeEventListener("pointercancel", i);
		};
		t.addEventListener("pointermove", r), t.addEventListener("pointerup", i), t.addEventListener("pointercancel", i);
	}
	function f(e) {
		e.key === "ArrowUp" ? (e.preventDefault(), to(eo.h + 24)) : e.key === "ArrowDown" && (e.preventDefault(), to(eo.h - 24));
	}
	var p = Ar(), m = un(p), h = (e) => {
		var o = To();
		let s, p;
		var m = R(o), h = (e) => {
			var t = vo();
			B(() => Z(t, "aria-valuenow", eo.h)), yr("pointerdown", t, d), yr("dblclick", t, function(...e) {
				no?.apply(this, e);
			}), yr("keydown", t, f), Y(e, t);
		};
		X(m, (e) => {
			q(n) && !eo.mobile && e(h);
		});
		var g = z(m, 2), _ = R(g);
		oi(_, 21, () => $.list, (e) => e.id, (e, r) => {
			var i = yo();
			let a;
			var o = R(i), s = dn(o, !0), u = z(o, 2);
			_o(R(u), { size: 12 }), k(u), k(i), B((e, t, o) => {
				a = wi(i, 1, "tab svelte-qk9xya", null, a, { on: q(r).id === q(n) }), Z(i, "aria-selected", q(r).id === q(n)), Z(i, "data-session", q(r).id), Z(i, "title", e), Gr(s, t), Z(u, "aria-label", o);
			}, [
				() => l(q(r).id),
				() => c(q(r).id),
				() => "Close tab " + c(q(r).id)
			]), yr("click", o, () => q(r).id === q(n) && !eo.mobile ? t.api.closeSession() : oo(t.api, q(r))), yr("click", u, () => so(t.api, q(n), q(r).id)), Y(e, i);
		}), k(_);
		var v = z(_, 2), y = (e) => {
			var n = bo();
			uo(R(n), { size: 16 }), k(n), yr("click", n, () => t.api.closeSession()), Y(e, n);
		}, b = (e) => {
			var n = xo();
			mo(R(n), { size: 16 }), k(n), yr("click", n, () => oo(t.api, $.list.at(-1) ?? $.list[0])), Y(e, n);
		};
		X(v, (e) => {
			q(n) ? e(y) : e(b, -1);
		}), k(g);
		var x = z(g, 2), S = (e) => {
			var o = wo(), s = R(o), c = (e) => {
				var o = Ar();
				ti(un(o), () => q(n) + q(r) + q(i), (e) => {
					var o = So();
					mi(o, (e, t) => oa?.(e, t), () => ({
						api: t.api,
						slot: "session",
						target: {
							plugin: q(a).plugin,
							element: q(a).element
						},
						context: {
							session: q(n),
							item: q(r),
							readonly: q(i)
						}
					})), B(() => Z(o, "data-plugin", q(a).plugin)), Y(e, o);
				}), Y(e, o);
			}, l = (e) => {
				Y(e, Co());
			};
			X(s, (e) => {
				q(a) ? e(c) : e(l, -1);
			}), k(o), Y(e, o);
		};
		X(x, (e) => {
			q(n) && e(S);
		}), k(o), Xi(o, (e) => F(u, e), () => q(u)), B(() => {
			s = wi(o, 1, "console svelte-qk9xya", null, s, { open: !!q(n) }), Z(o, "data-testid", q(n) ? "session-panel" : "console"), p = Ei(o, "", p, { "--ch": eo.h + "px" });
		}), Y(e, o);
	};
	X(m, (e) => {
		(q(n) || $.list.length && !eo.mobile) && e(h);
	}), Y(e, p), $e();
}
br([
	"pointerdown",
	"dblclick",
	"keydown",
	"click"
]);
// kit/format.js
function Do(e) {
	return !e || !e.fields ? {} : typeof e.fields == "string" ? JSON.parse(e.fields) : e.fields;
}
// plugins/desk/Drawer.svelte
var Oo = /* @__PURE__ */ J("<div class=\"err svelte-gsmuzk\"> </div>"), ko = /* @__PURE__ */ J("<div class=\"plugin-host\"></div>"), Ao = /* @__PURE__ */ J("<div class=\"row actions svelte-gsmuzk\" data-testid=\"drawer-actions\"></div>"), jo = /* @__PURE__ */ J("<h3 class=\"svelte-gsmuzk\"> </h3>"), Mo = /* @__PURE__ */ J("<!> <div data-testid=\"drawer-plugin\"><div class=\"plugin-host\"></div></div>", 1), No = /* @__PURE__ */ J("<p class=\"faint\">No drawer plugins are enabled.</p>"), Po = /* @__PURE__ */ J("<!> <!>", 1), Fo = /* @__PURE__ */ J("<p class=\"faint\">Loading…</p>"), Io = /* @__PURE__ */ J("<div class=\"backdrop svelte-gsmuzk\" role=\"presentation\"></div> <aside class=\"drawer svelte-gsmuzk\" data-testid=\"drawer\" aria-label=\"ticket\"><header class=\"svelte-gsmuzk\"><span class=\"mono muted id svelte-gsmuzk\"> </span> <button class=\"ghost icon\" aria-label=\"close\"><!></button></header> <div class=\"scrollbody svelte-gsmuzk\"><!></div></aside>", 1);
function Lo(e, t) {
	Qe(t, !0);
	let n = ia(() => t.api, () => "api/ticket/" + encodeURIComponent(t.id)), r = () => t.api.go({ ticket: "" }), i = (e, t) => ({
		kind: e.kind,
		label: e.labels ?? [],
		pool: e.pool || "",
		workflow: Do(t?.ticket ?? e).workflow || ""
	}), a = (e, n, r) => t.api.contributions(e, i(n, r)).sort((e, t) => Number(t.source === "builtin") - Number(e.source === "builtin")), o = (e) => Ke({
		ticket: e.ticket,
		run: e.run,
		trace: e.trace,
		comments: e.comments,
		sessions: e.sessions
	});
	var s = Io();
	vr("keydown", rn, (e) => {
		e.key === "Escape" && (e.preventDefault(), r());
	});
	var c = un(s), l = z(c, 2), u = R(l), d = R(u), f = dn(d, !0), p = z(d, 2);
	_o(R(p), { size: 14 }), k(p), k(u);
	var m = z(u, 2), h = R(m), g = (e) => {
		var t = Oo(), r = dn(t, !0);
		B(() => Gr(r, n.error)), Y(e, t);
	}, _ = (e) => {
		let r = /* @__PURE__ */ M(() => n.data.ticket), i = /* @__PURE__ */ M(() => o(n.data)), s = /* @__PURE__ */ M(() => a("action", q(r), n.data.run));
		var c = Po(), l = un(c), u = (e) => {
			var n = Ao();
			oi(n, 21, () => q(s), (e) => e.plugin + e.element, (e, n) => {
				var r = ko();
				mi(r, (e, t) => oa?.(e, t), () => ({
					api: t.api,
					slot: "action",
					target: {
						plugin: q(n).plugin,
						element: q(n).element
					},
					context: q(i)
				})), B(() => Z(r, "data-plugin", q(n).plugin)), Y(e, r);
			}), k(n), Y(e, n);
		};
		X(l, (e) => {
			q(s).length && e(u);
		}), oi(z(l, 2), 17, () => a("drawer", q(r), n.data.run), (e) => e.plugin + e.element, (e, n) => {
			var r = Mo(), a = un(r), o = (e) => {
				var t = jo(), r = dn(t, !0);
				B(() => Gr(r, q(n).title)), Y(e, t);
			};
			X(a, (e) => {
				q(n).title && e(o);
			});
			var s = z(a, 2), c = R(s);
			mi(c, (e, t) => oa?.(e, t), () => ({
				api: t.api,
				slot: "drawer",
				target: {
					plugin: q(n).plugin,
					element: q(n).element
				},
				context: q(i)
			})), k(s), B(() => {
				Z(s, "data-plugin", q(n).plugin), Z(c, "data-plugin", q(n).plugin);
			}), Y(e, r);
		}, (e) => {
			Y(e, No());
		}), Y(e, c);
	}, v = (e) => {
		Y(e, Fo());
	};
	X(h, (e) => {
		n.error ? e(g) : n.data ? e(_, 1) : e(v, -1);
	}), k(m), k(l), B(() => Gr(f, t.id)), yr("click", c, r), yr("click", p, r), Y(e, s), $e();
}
br(["click"]);
// plugins/desk/Frame.svelte
var Ro = /* @__PURE__ */ J("<div class=\"user faint svelte-e6uu72\" title=\"signed in through the proxy\"> </div>"), zo = /* @__PURE__ */ J("<span class=\"tabdot svelte-e6uu72\"></span>"), Bo = /* @__PURE__ */ J("<span class=\"badge svelte-e6uu72\"> </span>"), Vo = /* @__PURE__ */ J("<a><span class=\"ico svelte-e6uu72\"><!><!></span> <span class=\"lbl svelte-e6uu72\"> </span> <!></a>"), Ho = /* @__PURE__ */ J("<div class=\"hdr plugin-host svelte-e6uu72\"></div>"), Uo = /* @__PURE__ */ J("<div class=\"plugin-host\"></div>"), Wo = /* @__PURE__ */ J("<p class=\"faint pad\" data-testid=\"view-missing\">No plugin shows <span class=\"mono\"> </span>.</p>"), Go = /* @__PURE__ */ J("<div class=\"shell svelte-e6uu72\" data-testid=\"desk-app\"><div class=\"brand svelte-e6uu72\"><span class=\"mark svelte-e6uu72\"></span>Jobshop <span class=\"faint\">Desk</span></div> <nav aria-label=\"views\" class=\"svelte-e6uu72\"><!> <!> <!></nav> <div class=\"foot svelte-e6uu72\"><span class=\"live svelte-e6uu72\"><span></span> <span class=\"faint\"> </span></span> <span class=\"tools svelte-e6uu72\"><button class=\"ghost icon svelte-e6uu72\" aria-label=\"toggle theme\" title=\"toggle theme\"><!></button></span></div> <div class=\"col svelte-e6uu72\"><main class=\"svelte-e6uu72\"><!></main> <!></div> <!></div>");
function Ko(e, t) {
	Qe(t, !0);
	let n = na(t, "api", 7), r = aa(() => n()), i = /* @__PURE__ */ P(""), a = /* @__PURE__ */ P(!1), o = /* @__PURE__ */ P(en({})), s = /* @__PURE__ */ P(!1);
	$r(() => {
		F(s, n().theme === "light"), F(i, n().user, !0), q(i) || n().get("api/whoami").then((e) => F(i, e.user || "", !0)).catch(() => {});
		let e = n().live((e) => F(a, e, !0)), t = n().badges((e) => F(o, e, !0));
		return () => {
			e(), t();
		};
	});
	function c() {
		F(s, !q(s)), n().theme = q(s) ? "light" : "dark";
	}
	let l = {
		attention: {
			icon: ya,
			key: "a"
		},
		chat: {
			icon: Sa,
			key: "c"
		},
		work: {
			icon: Ta,
			key: "w"
		},
		activity: {
			icon: Oa,
			key: "y"
		},
		sessions: {
			icon: ja,
			key: "s"
		},
		pools: {
			icon: Pa,
			key: "p"
		},
		deploys: {
			icon: La,
			key: "d"
		}
	}, u = Object.keys(l), d = (e) => {
		let t = u.indexOf(e);
		return t < 0 ? u.length : t;
	}, f = /* @__PURE__ */ M(() => n().contributions("view")), p = /* @__PURE__ */ M(() => q(f).filter((e, t) => e.id && q(f).findIndex((t) => t.id === e.id) === t).map((e, t) => ({
		id: e.id ?? "",
		title: e.title || e.id || "",
		i: t
	})).sort((e, t) => d(e.id) - d(t.id) || e.i - t.i).map((e) => ({
		...e,
		icon: l[e.id]?.icon || Ka,
		key: l[e.id]?.key || ""
	}))), m = /* @__PURE__ */ M(() => n().contributions("header")), h = /* @__PURE__ */ M(() => r.view || "attention"), g = /* @__PURE__ */ M(() => q(f).find((e) => e.id === q(h))), _ = (e) => n().go(e), v = 0;
	function y(e) {
		if (e.key === "Escape" && r.session && !r.ticket && !e.defaultPrevented) {
			n().closeSession();
			return;
		}
		let t = e.target;
		if (!(e.metaKey || e.ctrlKey || e.altKey || t.closest("input, textarea, select, [contenteditable]"))) {
			if (v && Date.now() - v < 1200) {
				v = 0;
				let t = q(p).find((t) => t.key && t.key === e.key);
				t && (e.preventDefault(), _({
					view: t.id,
					ticket: ""
				}));
			} else e.key === "g" && (v = Date.now());
		}
	}
	var b = Go();
	vr("keydown", rn, y);
	var x = z(R(b), 2), S = R(x), C = (e) => {
		var t = Ro(), n = dn(t, !0);
		B(() => Gr(n, q(i))), Y(e, t);
	};
	X(S, (e) => {
		q(i) && e(C);
	});
	var w = z(S, 2);
	oi(w, 17, () => q(p), (e) => e.id, (e, t) => {
		let n = /* @__PURE__ */ M(() => q(t).icon);
		var r = Vo();
		let i;
		var a = R(r), s = R(a);
		fi(s, () => q(n), (e, t) => {
			t(e, { size: 14 });
		});
		var c = z(s), l = (e) => {
			Y(e, zo());
		};
		X(c, (e) => {
			(q(o)[q(t).id] ?? 0) > 0 && e(l);
		}), k(a);
		var u = z(a, 2), d = dn(u, !0), f = z(u, 2), p = (e) => {
			var n = Bo(), r = dn(n, !0);
			B(() => Gr(r, q(o)[q(t).id])), Y(e, n);
		};
		X(f, (e) => {
			(q(o)[q(t).id] ?? 0) > 0 && e(p);
		}), k(r), B(() => {
			Z(r, "href", "?view=" + q(t).id), Z(r, "data-view", q(t).id), Z(r, "title", "Go to " + q(t).title + (q(t).key ? " (g " + q(t).key + ")" : "")), i = wi(r, 1, "svelte-e6uu72", null, i, { active: q(h) === q(t).id }), Gr(d, q(t).title);
		}), yr("click", r, (e) => {
			e.preventDefault(), _({
				view: q(t).id,
				ticket: ""
			});
		}), Y(e, r);
	}), oi(z(w, 2), 17, () => q(m), (e) => e.plugin + e.element, (e, t) => {
		var r = Ho();
		mi(r, (e, t) => oa?.(e, t), () => ({
			api: n(),
			slot: "header",
			target: {
				plugin: q(t).plugin,
				element: q(t).element
			}
		})), B(() => Z(r, "data-plugin", q(t).plugin)), Y(e, r);
	}), k(x);
	var ee = z(x, 2), te = R(ee), ne = R(te);
	let re;
	var ie = dn(z(ne, 2), !0);
	k(te);
	var ae = z(te, 2), oe = R(ae), se = R(oe), ce = (e) => {
		Ua(e, { size: 14 });
	}, le = (e) => {
		Ba(e, { size: 14 });
	};
	X(se, (e) => {
		q(s) ? e(ce) : e(le, -1);
	}), k(oe), k(ae), k(ee);
	var ue = z(ee, 2), T = R(ue), de = R(T), fe = (e) => {
		var t = Ar();
		ti(un(t), () => q(g).plugin + q(g).element, (e) => {
			var t = Uo();
			mi(t, (e, t) => oa?.(e, t), () => ({
				api: n(),
				slot: "view",
				target: {
					plugin: q(g).plugin,
					element: q(g).element
				}
			})), B(() => Z(t, "data-plugin", q(g).plugin)), Y(e, t);
		}), Y(e, t);
	}, pe = (e) => {
		var t = Wo(), n = dn(z(R(t)), !0);
		De(), k(t), B(() => Gr(n, q(h))), Y(e, t);
	};
	X(de, (e) => {
		q(g) ? e(fe) : e(pe, -1);
	}), k(T), Eo(z(T, 2), {
		get api() {
			return n();
		},
		get url() {
			return r;
		}
	}), k(ue);
	var me = z(ue, 2), he = (e) => {
		Lo(e, {
			get api() {
				return n();
			},
			get id() {
				return r.ticket;
			}
		});
	};
	X(me, (e) => {
		r.ticket && e(he);
	}), k(b), B(() => {
		Z(te, "title", q(a) ? "live" : "disconnected"), re = wi(ne, 1, "dot", null, re, {
			ok: q(a),
			bad: !q(a)
		}), Gr(ie, q(a) ? "live" : "offline");
	}), yr("click", oe, c), Y(e, b), $e();
}
// plugins/desk/index.js
br(["click"]), ra("desk-app", Ko);
