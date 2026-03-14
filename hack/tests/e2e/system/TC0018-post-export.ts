import { test, expect } from '../../fixtures/auth';
import { PostPage } from '../../pages/PostPage';

test.describe('TC0018 岗位导出', () => {
  test('TC0018a: 导出岗位请求发送正确', async ({ adminPage }) => {
    const postPage = new PostPage(adminPage);
    await postPage.goto();

    // Set up request/response interception for the export API
    const responsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/api/post/export'),
      { timeout: 15000 },
    );

    await postPage.clickExport();
    const response = await responsePromise;

    // Verify the export request was sent and returned successfully
    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toContain('spreadsheetml');
  });
});
