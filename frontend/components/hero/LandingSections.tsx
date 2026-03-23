'use client';

import Link from 'next/link';
import { motion } from 'framer-motion';
import { ArrowRight, BarChart3, Boxes, Clapperboard, PlayCircle, ScanText, Sparkles } from 'lucide-react';
import PosterCarousel from '@/components/hero/PosterCarousel';
import { SiteCard, SiteSection, SiteSectionTitle, SiteIconBox } from '@/components/common/SitePrimitives';
import MagneticButton from '@/components/common/MagneticButton';

const PIPELINE = [
  { title: '脚本解析', meta: '章节与语义理解', icon: ScanText },
  { title: '镜头拆解', meta: '角色 · 场景 · 节奏', icon: Boxes },
  { title: '视频生成', meta: '多模型并行渲染', icon: Clapperboard },
  { title: '运营分发', meta: '账号与数据回流', icon: BarChart3 },
];

const resultBoard = [
  { label: '上线团队', value: '12+', hint: '影视解说 / 短剧 / 广告' },
  { label: '周产能峰值', value: '680 条', hint: '多模型并发生成' },
  { label: '平均交付', value: '47 分钟', hint: '脚本到成片' },
  { label: '运营回流', value: 'T+1', hint: '次日复盘' },
];

export default function LandingSections({ onRegister }: { onRegister: () => void }) {
  return (
    <div className="relative overflow-hidden bg-[#f7f7f7]">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_50%_-18%,rgba(0,0,0,0.05),transparent_45%)]" />

      {/* 第二屏：Tesla 风格横向大卡（左文右图） */}
      <SiteSection screen className="relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.25 }}
          transition={{ duration: 0.45 }}
        >
          <SiteCard className="site-card-strong site-sheen" padding="p-6 sm:p-8 lg:p-10">
            <div className="grid gap-8 lg:grid-cols-[0.92fr_1.08fr] lg:items-center">
              <div>
                <p className="site-kicker">Featured</p>
                <h2 className="site-h2 mt-4 text-balance">
                  一张横向大卡，
                  <span className="block text-black/55">聚焦核心能力与价值</span>
                </h2>
                <p className="site-lead mt-5 max-w-md text-black/60">
                  用案例画面做主视觉，文案只保留核心信息，避免左侧文字墙和信息噪声。
                </p>

                <div className="mt-7 grid grid-cols-2 gap-3 sm:max-w-md">
                  {['海报轮播', '流程拆解', '结果看板', '渠道分发'].map((item, idx) => (
                    <div
                      key={item}
                      className="rounded-xl border border-black/10 bg-black/[0.02] px-4 py-3 text-sm font-medium text-black/85"
                    >
                      <span className="inline-flex items-center ui-gap-8">
                        <Sparkles className="h-3.5 w-3.5 text-black/45" strokeWidth={1.35} />
                        {idx + 1}. {item}
                      </span>
                    </div>
                  ))}
                </div>

                <div className="mt-8">
                  <MagneticButton
                    onClick={onRegister}
                    className="btn-base btn-dark btn-m"
                  >
                    立即申请试用
                    <ArrowRight className="h-4 w-4" strokeWidth={1.5} />
                  </MagneticButton>
                </div>
              </div>

              <div className="lg:pl-2">
                <PosterCarousel />
              </div>
            </div>
          </SiteCard>
        </motion.div>
      </SiteSection>

      {/* 生产流程：四步独立卡片 */}
      <SiteSection screen className="relative z-10 border-t border-black/8 bg-white">
        <div className="w-full">
          <SiteSectionTitle
            centered
            overline="生产流程"
            title="从输入到复盘，一步一模块"
            subtitle="四段链路独立成卡，结构清晰，便于对外讲解与对内协作。"
          />

          <div className="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {PIPELINE.map((item, index) => (
              <motion.div
                key={item.title}
                initial={{ opacity: 0, y: 10 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true, amount: 0.25 }}
                transition={{ duration: 0.35, delay: index * 0.04 }}
              >
                <SiteCard padding="p-5" className="h-full">
                  <div className="flex items-start justify-between gap-3">
                    <SiteIconBox>
                      <item.icon />
                    </SiteIconBox>
                    <span className="text-xs tabular-nums text-black/35">0{index + 1}</span>
                  </div>
                  <h3 className="mt-4 text-base font-semibold text-black">{item.title}</h3>
                  <p className="mt-2 text-sm leading-relaxed text-black/55">{item.meta}</p>
                </SiteCard>
              </motion.div>
            ))}
          </div>
        </div>
      </SiteSection>

      {/* 数据看板：Bento 2×2 */}
      <SiteSection screen className="relative z-10">
        <div className="grid w-full gap-10 lg:grid-cols-12 lg:gap-12 lg:items-start">
          <div className="lg:col-span-5">
            <SiteSectionTitle
              overline="结果看板"
              title={
                <>
                  用数字说话，
                  <span className="block text-black/55">不用长文案堆砌</span>
                </>
              }
              subtitle="关键指标独立成块，扫一眼即可判断产能与效率。"
            />
          </div>
          <div className="grid grid-cols-2 gap-3 sm:gap-4 lg:col-span-7">
            {resultBoard.map((item) => (
              <SiteCard key={item.label} padding="p-5 sm:p-6" className="min-h-[140px]">
                <p className="text-xs font-medium uppercase tracking-wider text-black/45">{item.label}</p>
                <p className="site-stat-value mt-3">{item.value}</p>
                <p className="mt-2 text-xs leading-relaxed text-black/55">{item.hint}</p>
              </SiteCard>
            ))}
          </div>
        </div>
      </SiteSection>

      {/* CTA：窄幅居中 */}
      <SiteSection screen className="relative z-10 border-t border-black/8" tight>
        <div className="w-full">
          <SiteCard className="site-card-strong mx-auto max-w-2xl text-center" padding="p-10 sm:p-12">
            <p className="site-kicker">Get Started</p>
            <h3 className="site-h2 mt-4">预约演示或开通试用</h3>
            <p className="site-lead mx-auto mt-4 max-w-md text-black/60">
              完善海报素材与定价模块后，可与此区块衔接，形成完整商业闭环。
            </p>
            <div className="mt-8 flex flex-col items-stretch justify-center gap-3 sm:flex-row sm:items-center">
              <button
                type="button"
                onClick={onRegister}
                className="btn-base btn-dark btn-m"
              >
                <PlayCircle className="h-4 w-4" strokeWidth={1.5} />
                立即申请试用
              </button>
              <Link
                href="/factory/models"
                className="btn-base btn-light btn-m"
              >
                查看模型能力
                <ArrowRight className="h-4 w-4" strokeWidth={1.5} />
              </Link>
            </div>
          </SiteCard>
        </div>
      </SiteSection>
    </div>
  );
}
