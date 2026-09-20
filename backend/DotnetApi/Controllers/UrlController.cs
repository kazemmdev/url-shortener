namespace DotnetApi.Controllers;

using DotnetApi.Services;
using Microsoft.AspNetCore.Mvc;

[ApiController]
[Route("api/[controller]")]
public class UrlController(IUrlService service): ControllerBase
{
    [HttpPost]
    public async Task<IActionResult> Create([FromBody] UrlRequest request)
    {
        var shortUrl = await service.Create(request.LongUrl);
        return Ok(shortUrl);
    }

    [HttpGet("{shortCode}")]
    public async Task<IActionResult> GetByShortCode(string shortCode)
    {
        var longUrl = await service.GetByShortCode(shortCode);
        return longUrl == null ? NotFound() : Redirect(longUrl);
    }
}

