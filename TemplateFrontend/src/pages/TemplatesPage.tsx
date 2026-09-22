import { useCallback, useEffect, useMemo, useState } from 'react';
import { Download, FileText } from 'lucide-react';
import { downloadTemplate, listTemplates } from '../api/templates';
import { fetchPurchasedTemplateIds } from '../api/orders';
import { simulatePayment } from '../api/payment';
import { useAuth } from '../context/AuthContext';
import { ApiError } from '../api/client';
import type { DownloadData, Template } from '../types';
import { formatFileSize, formatPrice } from '../types';

interface PayModalState extends DownloadData {
  templateId: number;
  templateName: string;
}

const THUMB_THEMES = [
  { thumb: 'thumb-primary', dl: 'dl-primary', badge: 'badge-paid' },
  { thumb: 'thumb-coral', dl: 'dl-coral', badge: 'badge-paid' },
  { thumb: 'thumb-sage', dl: 'dl-sage', badge: 'badge-free' },
  { thumb: 'thumb-butter', dl: 'dl-butter', badge: 'badge-paid' },
  { thumb: 'thumb-periwinkle', dl: 'dl-periwinkle', badge: 'badge-paid' },
] as const;

function openDownload(url: string) {
  if (!url) return;
  window.open(url, '_blank', 'noopener,noreferrer');
}

async function waitForDownloadReady(
  token: string,
  templateId: number,
  maxAttempts = 10,
): Promise<DownloadData | null> {
  for (let i = 0; i < maxAttempts; i++) {
    await new Promise((r) => setTimeout(r, 600));
    const result = await downloadTemplate(token, templateId);
    if (result.order_status === 2 && result.download_url) {
      return result;
    }
  }
  return null;
}

export function TemplatesPage() {
  const { accessToken } = useAuth();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [actionId, setActionId] = useState<number | null>(null);
  const [payModal, setPayModal] = useState<PayModalState | null>(null);
  const [paying, setPaying] = useState(false);
  const [purchasedIds, setPurchasedIds] = useState<Set<number>>(new Set());
  const [typeFilter, setTypeFilter] = useState('all');
  const [priceFilter, setPriceFilter] = useState('all');

  const pageSize = 12;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const fetchPurchased = useCallback(async () => {
    if (!accessToken) return;
    try {
      const ids = await fetchPurchasedTemplateIds(accessToken);
      setPurchasedIds(ids);
    } catch {
      // 不影响模板列表展示
    }
  }, [accessToken]);

  const fetchTemplates = useCallback(async () => {
    if (!accessToken) return;
    setLoading(true);
    setError('');
    try {
      const data = await listTemplates(accessToken, page, pageSize);
      setTemplates(data.templates ?? []);
      setTotal(data.total ?? 0);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '加载模板失败');
    } finally {
      setLoading(false);
    }
  }, [accessToken, page]);

  useEffect(() => {
    fetchTemplates();
  }, [fetchTemplates]);

  useEffect(() => {
    fetchPurchased();
  }, [fetchPurchased]);

  const filteredTemplates = useMemo(() => {
    return templates.filter((t) => {
      if (typeFilter !== 'all' && t.file_type !== typeFilter) return false;
      if (priceFilter === 'free' && t.is_free !== 1) return false;
      if (priceFilter === 'paid' && t.is_free === 1) return false;
      return true;
    });
  }, [templates, typeFilter, priceFilter]);

  const fileTypes = useMemo(() => {
    const types = new Set(templates.map((t) => t.file_type));
    return Array.from(types).sort();
  }, [templates]);

  async function handleDownload(template: Template) {
    if (!accessToken) return;
    setActionId(template.id);
    setMessage('');
    setError('');

    try {
      const result = await downloadTemplate(accessToken, template.id);

      if (result.order_status === 2 && result.download_url) {
        openDownload(result.download_url);
        setMessage(`「${template.name}」下载已开始`);
        return;
      }

      if (result.order_status === 1 && result.payment_no) {
        setPayModal({
          ...result,
          templateId: template.id,
          templateName: template.name,
        });
        return;
      }

      setError('暂无法下载，请稍后重试');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '下载失败');
    } finally {
      setActionId(null);
    }
  }

  async function handlePay() {
    if (!payModal || !accessToken) return;
    setPaying(true);
    setError('');
    setMessage('');

    try {
      await simulatePayment(payModal.payment_no, payModal.price_snapshot);
      setMessage('支付成功，正在等待订单更新…');

      const ready = await waitForDownloadReady(accessToken, payModal.templateId);
      setPayModal(null);
      setPurchasedIds((prev) => new Set(prev).add(payModal.templateId));

      if (ready?.download_url) {
        openDownload(ready.download_url);
        setMessage(`「${payModal.templateName}」支付完成，下载已开始`);
      } else {
        setMessage('支付已完成，请再次点击下载按钮获取文件');
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '支付失败');
    } finally {
      setPaying(false);
    }
  }

  return (
    <div className="page">
      <div className="page-head-row">
        <div>
          <h2>全部模板</h2>
          <p>浏览并下载精选办公模板</p>
        </div>
        <div className="filter-row">
          <label className="filter-pill">
            类型
            <select value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)}>
              <option value="all">全部</option>
              {fileTypes.map((ft) => (
                <option key={ft} value={ft}>
                  {ft}
                </option>
              ))}
            </select>
          </label>
          <label className="filter-pill">
            定价
            <select value={priceFilter} onChange={(e) => setPriceFilter(e.target.value)}>
              <option value="all">全部</option>
              <option value="free">免费</option>
              <option value="paid">付费</option>
            </select>
          </label>
          <button
            type="button"
            className="btn btn-ghost"
            onClick={() => {
              fetchTemplates();
              fetchPurchased();
            }}
          >
            刷新
          </button>
        </div>
      </div>

      {message && <div className="alert alert-success">{message}</div>}
      {error && <div className="alert alert-error">{error}</div>}

      {loading ? (
        <div className="state-box">加载中…</div>
      ) : filteredTemplates.length === 0 ? (
        <div className="state-box">暂无符合条件的模板</div>
      ) : (
        <div className="template-grid">
          {filteredTemplates.map((t, index) => {
            const theme = THUMB_THEMES[index % THUMB_THEMES.length];
            const badgeClass = t.is_free === 1 ? 'badge-free' : theme.badge;
            return (
              <article key={t.id} className="template-card">
                <div className={`template-thumb ${theme.thumb}`}>
                  {t.thumbnail_download_url ? (
                    <img src={t.thumbnail_download_url} alt={t.name} loading="lazy" />
                  ) : (
                    <div className="thumb-inner">
                      <FileText size={32} />
                    </div>
                  )}
                </div>
                <div className="template-body">
                  <h3>{t.name}</h3>
                  <p className="template-desc">{t.description || '暂无描述'}</p>
                  <div className="template-meta">
                    <div className="template-badges">
                      <span className={`badge ${badgeClass}`}>
                        {t.is_free === 1 ? '免费' : formatPrice(t.price)}
                      </span>
                      {t.is_free !== 1 && purchasedIds.has(t.id) && (
                        <span className="badge badge-purchased">已购买</span>
                      )}
                    </div>
                    <span className="meta-text">
                      {t.file_type} · {formatFileSize(t.file_size)}
                    </span>
                  </div>
                  <div className="tpl-footer">
                    <span className="meta-text">点击右侧下载</span>
                    <button
                      type="button"
                      className={`btn dl-btn ${theme.dl}`}
                      disabled={actionId === t.id}
                      aria-label={`下载 ${t.name}`}
                      onClick={() => handleDownload(t)}
                    >
                      <Download size={18} />
                    </button>
                  </div>
                </div>
              </article>
            );
          })}
        </div>
      )}

      {totalPages > 1 && (
        <div className="pagination">
          <button
            type="button"
            className="btn btn-ghost"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
          >
            上一页
          </button>
          <span>
            第 {page} / {totalPages} 页 · 共 {total} 个
          </span>
          <button
            type="button"
            className="btn btn-ghost"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
          >
            下一页
          </button>
        </div>
      )}

      {payModal && (
        <div className="modal-overlay" role="presentation" onClick={() => !paying && setPayModal(null)}>
          <div
            className="modal-card"
            role="dialog"
            aria-modal="true"
            onClick={(e) => e.stopPropagation()}
          >
            <h3>模拟支付</h3>
            <p>
              模板「{payModal.templateName}」需支付{' '}
              <strong>{formatPrice(payModal.price_snapshot)}</strong>
            </p>
            <dl className="pay-details">
              <dt>订单号</dt>
              <dd>{payModal.order_no}</dd>
              <dt>支付单号</dt>
              <dd>{payModal.payment_no}</dd>
            </dl>
            <p className="pay-hint">点击下方按钮将调用 PayWebServer 模拟支付回调</p>
            <div className="modal-actions">
              <button
                type="button"
                className="btn btn-ghost"
                disabled={paying}
                onClick={() => setPayModal(null)}
              >
                取消
              </button>
              <button type="button" className="btn btn-primary" disabled={paying} onClick={handlePay}>
                {paying ? '支付中…' : '确认支付'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
