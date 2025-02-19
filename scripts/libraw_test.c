#include <stdio.h>
#include <libraw/libraw.h>

int main() {
    libraw_data_t *raw = libraw_init(0);
    libraw_output_params_t *params = &raw->params;

    printf("Greybox: %u %u %u %u\n", params->greybox[0], params->greybox[1], params->greybox[2], params->greybox[3]);
    printf("Cropbox: %u %u %u %u\n", params->cropbox[0], params->cropbox[1], params->cropbox[2], params->cropbox[3]);
    printf("Aberration Correction: %f %f %f %f\n", params->aber[0], params->aber[1], params->aber[2], params->aber[3]);
    printf("Gamma: %f %f %f %f %f %f\n", params->gamm[0], params->gamm[1], params->gamm[2], params->gamm[3], params->gamm[4], params->gamm[5]);
    printf("User Multipliers: %f %f %f %f\n", params->user_mul[0], params->user_mul[1], params->user_mul[2], params->user_mul[3]);
    printf("Brightness: %f\n", params->bright);
    printf("Threshold: %f\n", params->threshold);
    
    printf("Half Size: %d\n", params->half_size);
    printf("Four Color RGB: %d\n", params->four_color_rgb);
    printf("Highlight: %d\n", params->highlight);
    printf("Use Auto WB: %d\n", params->use_auto_wb);
    printf("Use Camera WB: %d\n", params->use_camera_wb);
    printf("Use Camera Matrix: %d\n", params->use_camera_matrix);
    
    printf("Output Color: %d\n", params->output_color);
    printf("Output Profile: %s\n", params->output_profile ? params->output_profile : "NULL");
    printf("Camera Profile: %s\n", params->camera_profile ? params->camera_profile : "NULL");
    printf("Bad Pixels: %s\n", params->bad_pixels ? params->bad_pixels : "NULL");
    printf("Dark Frame: %s\n", params->dark_frame ? params->dark_frame : "NULL");
    
    printf("Output BPS: %d\n", params->output_bps);
    printf("Output TIFF: %d\n", params->output_tiff);
    printf("Output Flags: %d\n", params->output_flags);
    
    printf("User Flip: %d\n", params->user_flip);
    printf("User Quality: %d\n", params->user_qual);
    printf("User Black: %d\n", params->user_black);
    printf("User CBlack: %d %d %d %d\n", params->user_cblack[0], params->user_cblack[1], params->user_cblack[2], params->user_cblack[3]);
    printf("User Saturation: %d\n", params->user_sat);
    
    printf("Median Filter Passes: %d\n", params->med_passes);
    printf("Auto Bright Threshold: %f\n", params->auto_bright_thr);
    printf("Adjust Maximum Threshold: %f\n", params->adjust_maximum_thr);
    
    printf("No Auto Bright: %d\n", params->no_auto_bright);
    printf("Use Fuji Rotate: %d\n", params->use_fuji_rotate);
    printf("Green Matching: %d\n", params->green_matching);
    
    printf("DCB Iterations: %d\n", params->dcb_iterations);
    printf("DCB Enhance FL: %d\n", params->dcb_enhance_fl);
    printf("FBDD Noise Reduction: %d\n", params->fbdd_noiserd);
    
    printf("Exposure Correction: %d\n", params->exp_correc);
    printf("Exposure Shift: %f\n", params->exp_shift);
    printf("Exposure Preservation: %f\n", params->exp_preser);
    
    printf("No Auto Scale: %d\n", params->no_auto_scale);
    printf("No Interpolation: %d\n", params->no_interpolation);
    
    libraw_close(raw);
    return 0;
}
