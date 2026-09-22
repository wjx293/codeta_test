import { FormEvent, useEffect, useState } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import { UploadCloud } from 'lucide-react';
import { uploadFile } from '../api/upload';
import {
  createTemplate,
  updateTemplate,
  type CreateTemplatePayload,
} from '../api/templates';
import { ApiError } from '../api/client';
import { getFileExtension, type Template } from '../types';

const TEMPLATE_TYPES = ['ppt', 'pptx', 'doc', 'docx'];
const MAX_TEMPLATE_SIZE = 5 * 1024 * 1024;
const MAX_THUMB_SIZE = 1 * 1024 * 1024;

interface FormState {
  name: string;
  description: string;
  isFree: boolean;
  priceYuan: string;
}

export function TemplateFormPage() {
  const { id } = useParams();
  const location = useLocation();
  const isEdit = Boolean(id);
  const navigate = useNavigate();
  const existing = (location.state as { template?: Template } | null)?.template;

  const [form, setForm] = useState<FormState>({
    name: '',
    description: '',
    isFree: true,
    priceYuan: '0',
  });
  const [templateFile, setTemplateFile] = useState<File | null>(null);
  const [thumbFile, setThumbFile] = useState<File | null>(null);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (isEdit && existing) {
      setForm({
        name: existing.name,
        description: existing.description,
        isFree: existing.is_free === 1,
        priceYuan: existing.is_free === 1 ? '0' : (existing.price / 100).toFixed(2),
      });
    } else if (isEdit && !existing) {
      setMessage('请从模板列表进入编辑页');
    }
  }, [isEdit, existing]);

  function validateFile(file: File, allowed: string[], maxSize: number, label: string) {
    const ext = getFileExtension(file.name);
    if (!allowed.includes(ext)) {
      throw new Error(`${label}格式不支持：${allowed.join('/')}`);
    }
    if (file.size > maxSize) {
      throw new Error(`${label}超过大小限制`);
    }
    return ext;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setMessage('');
    setSubmitting(true);

    try {
      const priceCents = form.isFree ? 0 : Math.round(parseFloat(form.priceYuan || '0') * 100);
      if (!form.isFree && priceCents <= 0) {
        throw new Error('付费模板价格必须大于 0');
      }

      let thumbnailOssKey = '';
      let thumbnailFileSize = 0;

      if (thumbFile) {
        validateFile(thumbFile, ['jpg', 'jpeg', 'png'], MAX_THUMB_SIZE, '缩略图');
        setMessage('正在上传缩略图…');
        thumbnailOssKey = await uploadFile(thumbFile, true);
        thumbnailFileSize = thumbFile.size;
      }

      if (isEdit && id) {
        await updateTemplate(Number(id), {
          name: form.name,
          description: form.description,
          is_free: form.isFree ? 1 : 0,
          price: priceCents,
          thumbnail_oss_key: thumbnailOssKey || undefined,
          thumbnail_file_size: thumbnailFileSize || undefined,
        });
        setMessage('模板已更新');
        navigate('/templates');
        return;
      }

      if (!templateFile) {
        throw new Error('请选择模板文件');
      }

      const fileType = validateFile(templateFile, TEMPLATE_TYPES, MAX_TEMPLATE_SIZE, '模板文件');

      setMessage('正在上传模板文件…');
      const fileOssKey = await uploadFile(templateFile, false);

      const payload: CreateTemplatePayload = {
        name: form.name,
        description: form.description,
        is_free: form.isFree ? 1 : 0,
        price: priceCents,
        file_type: fileType,
        file_size: templateFile.size,
        file_oss_key: fileOssKey,
        thumbnail_oss_key: thumbnailOssKey || undefined,
        thumbnail_file_size: thumbnailFileSize || undefined,
      };

      await createTemplate(payload);
      setMessage('模板创建成功');
      navigate('/templates');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : err instanceof Error ? err.message : '提交失败');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="page">
      <header className="dashboard-head">
        <div>
          <p className="dashboard-greet">{isEdit ? '修改模板信息与定价' : '填写模板信息并上传文件'}</p>
          <h2 className="dashboard-title">{isEdit ? '编辑模板' : '新建模板'}</h2>
        </div>
        <Link to="/templates" className="btn btn-ghost">
          返回列表
        </Link>
      </header>

      {message && <div className="alert alert-info">{message}</div>}
      {error && <div className="alert alert-error">{error}</div>}

      <form className="form-card" onSubmit={handleSubmit}>
        <div className="form-row">
          <label>
            模板名称 *
            <input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="输入模板名称"
              required
            />
          </label>
          {!isEdit && (
            <label>
              文件类型
              <select
                value={templateFile ? getFileExtension(templateFile.name) : ''}
                disabled
              >
                <option value="">上传后自动识别</option>
              </select>
            </label>
          )}
          {isEdit && existing && (
            <label>
              文件类型
              <input value={existing.file_type} readOnly />
            </label>
          )}
        </div>

        <label>
          描述
          <textarea
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            placeholder="简要描述模板特点"
            rows={4}
          />
        </label>

        <div className="form-row">
          <fieldset className="inline-fieldset">
            <legend>定价</legend>
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={form.isFree}
                onChange={(e) => setForm({ ...form, isFree: e.target.checked })}
              />
              免费模板
            </label>
            {!form.isFree && (
              <label>
                定价（元）
                <input
                  type="number"
                  min="0.01"
                  step="0.01"
                  value={form.priceYuan}
                  onChange={(e) => setForm({ ...form, priceYuan: e.target.value })}
                  placeholder="0 表示免费"
                  required
                />
              </label>
            )}
          </fieldset>
          {isEdit && existing && (
            <label>
              状态
              <input
                value={existing.status === 1 ? '已上架' : '已下架'}
                readOnly
              />
            </label>
          )}
        </div>

        {!isEdit && (
          <div className="upload-zone">
            <UploadCloud size={36} color="var(--primary)" />
            <strong>拖拽文件到此处，或点击选择</strong>
            <span>支持 .ppt / .pptx / .doc / .docx，最大 5MB</span>
            <input
              type="file"
              accept=".ppt,.pptx,.doc,.docx"
              onChange={(e) => setTemplateFile(e.target.files?.[0] ?? null)}
              required
            />
            {templateFile && <span>{templateFile.name}</span>}
          </div>
        )}

        <div className="upload-zone">
          <UploadCloud size={28} color="var(--periwinkle)" />
          <strong>缩略图（可选）</strong>
          <span>未上传时将尝试从模板文件内嵌封面自动提取（pptx/docx）</span>
          <input
            type="file"
            accept=".jpg,.jpeg,.png"
            onChange={(e) => setThumbFile(e.target.files?.[0] ?? null)}
          />
          {thumbFile && <span>{thumbFile.name}</span>}
        </div>

        <div className="form-actions">
          <button type="submit" className="btn btn-primary" disabled={submitting}>
            {submitting ? '提交中…' : isEdit ? '保存' : '创建模板'}
          </button>
          <Link to="/templates" className="btn btn-ghost">
            取消
          </Link>
        </div>
      </form>
    </div>
  );
}
