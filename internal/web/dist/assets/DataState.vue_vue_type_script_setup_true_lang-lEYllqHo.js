import{x as f,a1 as k,A as v,r as m,d as p,c as s,F as y,D as b,k as h,b as a,t as i,a8 as x,O as _,l as w,f as d,q as t,e as g,L as A}from"./index-zYeq6rGf.js";/**
 * @license lucide-vue-next v0.460.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const C=f("ChevronRightIcon",[["path",{d:"m9 18 6-6-6-6",key:"mthhwq"}]]);/**
 * @license lucide-vue-next v0.460.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const $=f("InboxIcon",[["polyline",{points:"22 12 16 12 14 15 10 15 8 12 2 12",key:"o97t9d"}],["path",{d:"M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z",key:"oot6mr"}]]);/**
 * @license lucide-vue-next v0.460.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const B=f("TriangleAlertIcon",[["path",{d:"m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3",key:"wmoenq"}],["path",{d:"M12 9v4",key:"juzpu7"}],["path",{d:"M12 17h.01",key:"p32p05"}]]);function V(e,c=!0){const n=m(null),r=m(!1),l=m(null);async function o(){r.value=!0,l.value=null;try{n.value=await e()}catch(u){l.value=(u instanceof v,u.message)}finally{r.value=!1}}return c&&k(o),{data:n,loading:r,error:l,reload:o}}const I={class:"mb-6 space-y-2.5"},j={key:0,class:"flex items-center gap-1 text-xs text-muted-foreground"},D={class:"flex items-end justify-between gap-4"},L={class:"text-[28px] font-semibold leading-tight tracking-tight text-foreground"},M={key:0,class:"mt-1 text-sm text-muted-foreground"},T={class:"flex shrink-0 items-center gap-2 pb-1"},F=p({__name:"PageHeader",props:{title:{},description:{},breadcrumb:{}},setup(e){return(c,n)=>{var r;return t(),s("header",I,[(r=e.breadcrumb)!=null&&r.length?(t(),s("nav",j,[(t(!0),s(y,null,b(e.breadcrumb,(l,o)=>(t(),s(y,{key:o},[a("span",{class:_(o===e.breadcrumb.length-1?"font-medium text-foreground":"")},i(l),3),o<e.breadcrumb.length-1?(t(),w(d(C),{key:0,class:"h-3 w-3"})):h("",!0)],64))),128))])):h("",!0),a("div",D,[a("div",null,[a("h1",L,i(e.title),1),e.description?(t(),s("p",M,i(e.description),1)):h("",!0)]),a("div",T,[x(c.$slots,"actions")])])])}}}),q={key:0,class:"flex items-center justify-center gap-2 py-16 text-muted-foreground"},z={key:1,class:"flex items-center gap-3 rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive"},E={key:2,class:"flex flex-col items-center justify-center gap-2 py-16 text-muted-foreground"},N={class:"text-sm"},H=p({__name:"DataState",props:{loading:{type:Boolean},error:{},empty:{type:Boolean},emptyText:{}},setup(e){return(c,n)=>e.loading?(t(),s("div",q,[g(d(A),{class:"h-5 w-5 animate-spin"}),n[0]||(n[0]=a("span",{class:"text-sm"},"Wird geladen …",-1))])):e.error?(t(),s("div",z,[g(d(B),{class:"h-5 w-5 shrink-0"}),a("span",null,i(e.error),1)])):e.empty?(t(),s("div",E,[g(d($),{class:"h-8 w-8"}),a("span",N,i(e.emptyText||"Keine Einträge vorhanden."),1)])):x(c.$slots,"default",{key:3})}});export{C,F as _,H as a,V as u};
