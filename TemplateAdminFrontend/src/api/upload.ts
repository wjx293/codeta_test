import OSS from 'ali-oss';
import { apiRequest } from './client';
import type { ConfirmUploadData, UploadCredential } from '../types';

export async function getUploadCredential(
  fileType: string,
  fileSize: number,
  isThumbnail: boolean,
): Promise<UploadCredential> {
  return apiRequest<UploadCredential>('/api/admin/upload/credential', {
    method: 'POST',
    body: { file_type: fileType, file_size: fileSize, is_thumbnail: isThumbnail },
  });
}

export async function confirmUpload(
  objectKey: string,
  fileSize: number,
): Promise<ConfirmUploadData> {
  return apiRequest<ConfirmUploadData>('/api/admin/upload/confirm', {
    method: 'POST',
    body: { object_key: objectKey, file_size: fileSize },
  });
}

function isMockCredential(cred: UploadCredential): boolean {
  return cred.access_key_id.startsWith('mock-');
}

/** 真实 OSS 上传必须使用 STS 临时三元组，禁止无 security_token 的永久密钥 */
function assertSTSCredential(cred: UploadCredential): void {
  if (isMockCredential(cred)) {
    return;
  }
  if (!cred.security_token) {
    throw new Error('invalid upload credential: STS security_token is required');
  }
}

/** 使用 ali-oss SDK + STS 临时凭证直传；mock 环境跳过真实上传 */
async function putToOss(file: File, cred: UploadCredential): Promise<void> {
  if (!cred.bucket || !cred.endpoint || isMockCredential(cred)) {
    return;
  }

  assertSTSCredential(cred);

  const region = cred.region.startsWith('oss-') ? cred.region : `oss-${cred.region}`;

  const client = new OSS({
    accessKeyId: cred.access_key_id,
    accessKeySecret: cred.access_key_secret,
    stsToken: cred.security_token,
    bucket: cred.bucket,
    region,
    secure: true,
  });

  await client.put(cred.object_key, file);
}

export async function uploadFile(file: File, isThumbnail: boolean): Promise<string> {
  const ext = file.name.includes('.') ? file.name.split('.').pop()!.toLowerCase() : 'bin';
  const cred = await getUploadCredential(ext, file.size, isThumbnail);
  await putToOss(file, cred);
  await confirmUpload(cred.object_key, file.size);
  return cred.object_key;
}
